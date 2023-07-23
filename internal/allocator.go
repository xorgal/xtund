package internal

import (
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"

	bolt "go.etcd.io/bbolt"
)

const (
	dbBucket = "allocator"
)

type Allocator struct {
	db   *bolt.DB
	mu   sync.Mutex
	cidr string
}

func CreateAllocator(cidr string) (*Allocator, error) {
	db, err := bolt.Open(FilePath.AllocatorPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	a := &Allocator{
		db:   db,
		cidr: cidr,
	}
	// Initialize the bucket
	err = a.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dbBucket))
		return err
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Allocator) RegisterDevice(id string) (string, string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var ip net.IP
	// Check the DB for the DeviceId
	err := a.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbBucket))
		ipBytes := bucket.Get([]byte(id))
		if ipBytes != nil {
			ip = bytesToIP(ipBytes)
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	// If IP was not found for DeviceId, generate a new IP
	if ip == nil {
		allocatedIP, err := a.generateIP()
		if err != nil {
			return "", "", err
		}
		// Save to the DB
		err = a.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(dbBucket))
			return bucket.Put([]byte(id), ipToBytes(allocatedIP))
		})
		if err != nil {
			return "", "", err
		}
		ip = allocatedIP
	}
	// Create a CIDR block
	cidr := ip.String() + "/" + strings.Split(a.cidr, "/")[1]
	serverIP := strings.Split(a.cidr, "/")[0]
	return cidr, serverIP, nil
}

func (a *Allocator) generateIP() (net.IP, error) {
	var ip net.IP
	err := a.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbBucket))
		serverIP, serverNet, err := net.ParseCIDR(a.cidr)
		if err != nil {
			return err
		}
		// Calculate the first and last IP in the range based on the mask
		ones, bits := serverNet.Mask.Size()
		// Start IP range from serverIP + 2 (to exclude server IP and network address)
		ipInt := big.NewInt(0).SetBytes(serverIP.To4())
		ipInt.Add(ipInt, big.NewInt(2))
		// Calculate the last IP as the highest host IP in the subnet
		hostPartMax := big.NewInt(0).Lsh(big.NewInt(1), uint(bits-ones))
		hostPartMax.Sub(hostPartMax, big.NewInt(2)) // subtract 2 to exclude broadcast and zero addresses
		// Convert server IP to big Int
		serverIPInt := big.NewInt(0).SetBytes(serverIP.To4())
		// Mask out the host part of server IP
		networkPart := big.NewInt(0).Lsh(big.NewInt(1), uint(ones))
		networkPart.Sub(networkPart, big.NewInt(1))
		networkPart.Lsh(networkPart, uint(bits-ones))
		networkPart.And(networkPart, serverIPInt)
		// Calculate last IP by adding network part and host part
		lastIPInt := big.NewInt(0).Or(networkPart, hostPartMax)
		lastIP := make(net.IP, net.IPv4len)
		paddedBytes := append(make([]byte, net.IPv4len-len(lastIPInt.Bytes())), lastIPInt.Bytes()...)
		copy(lastIP, paddedBytes)
		// Find the first available IP in the range
		for {
			ip = make(net.IP, net.IPv4len)
			copy(ip, ipInt.Bytes())
			// Check if the IP is already allocated
			if bucket.Get(ipToBytes(ip)) == nil {
				break
			}
			// Increment the IP by 1
			ipInt.Add(ipInt, big.NewInt(1))
			// Check if the IP exceeds the CIDR range
			if ipInt.Cmp(lastIPInt) > 0 {
				return fmt.Errorf("no available IP in the range")
			}
		}
		// Save the allocated IP to the DB
		err = bucket.Put(ipToBytes(ip), []byte("allocated"))
		if err != nil {
			return err
		}
		return nil
	})
	return ip, err
}

func ipToBytes(ip net.IP) []byte {
	return ip.To4()
}

func bytesToIP(ip []byte) net.IP {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3])
}
