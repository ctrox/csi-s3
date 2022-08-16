package common

import (
	"hash/fnv"
	"sync"

	"k8s.io/utils/mount"
)

type KeyMutex struct {
	mutexes []sync.RWMutex
	size    int32
}

func HashToUint32(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)

	return h.Sum32()
}

func NewKeyMutex(size int32) *KeyMutex {
	return &KeyMutex{
		mutexes: make([]sync.RWMutex, size),
		size:    size,
	}
}

func (km *KeyMutex) GetMutex(key string) *sync.RWMutex {
	hashed := HashToUint32([]byte(key))
	index := hashed % uint32(km.size)

	return &km.mutexes[index]
}

// CleanupMountPoint unmounts the given path and deletes the remaining directory
func CleanupMountPoint(mountPath string) error {
	mounter := mount.New("")

	if err := mount.CleanupMountPoint(mountPath, mounter, true); err != nil {
		return err
	}

	return nil
}
