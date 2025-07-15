package cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheEntry[V any] struct {
	value     V
	createdAt time.Time
	ttl       time.Duration // Time To Live
}

// CacheManager mengelola semua item cache
type CacheManager[V any] struct {
	data map[string]CacheEntry[V]
	mu   sync.RWMutex // Mutex untuk melindungi akses ke map
}

// NewCacheManager membuat instance CacheManager baru
func NewCacheManager[V any]() *CacheManager[V] {
	return &CacheManager[V]{
		data: make(map[string]CacheEntry[V]),
	}
}

// Remember berfungsi seperti cache remember di Laravel
// Mengambil nilai dari cache jika ada dan belum kadaluwarsa.
// Jika tidak ada atau sudah kadaluwarsa, memanggil fungsi `callback`
// untuk mendapatkan nilai, menyimpannya di cache, dan mengembalikannya.
func (cm *CacheManager[V]) Remember(key string, ttl time.Duration, callback func() (V, error)) (V, error) {
	// Baca kunci (read lock) untuk memeriksa keberadaan dan validitas cache
	cm.mu.RLock()
	entry, found := cm.data[key]
	cm.mu.RUnlock()
	var zeroValue V

	if found && time.Since(entry.createdAt) < entry.ttl {
		fmt.Printf("Cache hit for key: %s\n", key)
		return entry.value, nil // Cache masih valid
	}

	// Jika tidak ditemukan atau kadaluwarsa, dapatkan write lock
	// untuk memperbarui cache
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Periksa kembali setelah mendapatkan write lock (double-checked locking)
	// untuk menghindari race condition jika goroutine lain sudah mengupdate cache
	entry, found = cm.data[key]
	if found && time.Since(entry.createdAt) < entry.ttl {
		fmt.Printf("Cache hit (after recheck) for key: %s\n", key)
		return entry.value, nil
	}

	fmt.Printf("Cache miss for key: %s. Executing callback...\n", key)
	// Cache miss atau kadaluwarsa, panggil callback
	value, err := callback()
	if err != nil {
		return zeroValue, err
	}

	// Simpan nilai baru ke cache
	cm.data[key] = CacheEntry[V]{
		value:     value,
		createdAt: time.Now(),
		ttl:       ttl,
	}
	return value, nil
}

// Forget menghapus item dari cache
func (cm *CacheManager[V]) Forget(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.data, key)
	fmt.Printf("Cache forgotten for key: %s\n", key)
}
