package mapcache

import (
	"fmt"
	"testing"
	"time"
)

func init() {
	ttlInSecond = 1
}
func TestSingleton(t *testing.T) {
	inst1 := GetInstance()
	inst2 := GetInstance()
	if inst1 != inst2 {
		t.Error("Instances not equal")
	}
	t.Log(inst1, inst2)
}

func TestHelperGetDuration(t *testing.T) {
	td := getTimeDurationFunc(10)
	if td.String() != "10s" {
		t.Error("Time duration incorrect calculation")
	}
	t.Log(td)
}

func TestCacheItemSetAndExpire(t *testing.T) {
	cache := GetInstance()

	tt := []struct {
		caseName                string
		sleepTimeInMilliseconds time.Duration
		itemsToBeAdded          int
		cacheLengthExpected     int
	}{
		{"Add 10000 items concurrently", 20, 10000, 10000},
		{"Add 5000 more items", 20, 5000, 15000},
		{"Add 3000 after first batch expired", 950, 3000, 8000},
		{"Add 2000 after others expired", 980, 2000, 2000},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			for i := 0; i < tc.itemsToBeAdded; i++ {
				go cache.Set(fmt.Sprintf("%s%v", tc.caseName, i), i)
			}

			time.Sleep(tc.sleepTimeInMilliseconds * time.Millisecond)

			if len(instance.cache) != tc.cacheLengthExpected {
				t.Error("Cache length is incorrect")
			}
			t.Logf("%s: %v cache items", tc.caseName, len(instance.cache))
		})
	}

}

func TestCacheItemGetNotExisted(t *testing.T) {
	cache := GetInstance()

	_, ok := cache.Get("nokey")
	if ok {
		t.Error("Got something using unexisted key")
	}
}

func TestCacheItemGetAndDelete(t *testing.T) {
	cache := GetInstance()
	expectedValue := 100
	testKey := "akey"
	cache.Set(testKey, expectedValue)

	tt := []struct {
		caseName             string
		testKey              string
		timeoutForCancelFunc time.Duration
		ok                   bool
	}{
		{"Get an item", testKey, 0, true},
		{"Get an item before delete", testKey, 500, true},
		{"Get an item after delete", testKey, 0, false},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			_, ok := cache.Get(tc.testKey)
			time.Sleep(tc.timeoutForCancelFunc * time.Millisecond)
			if ok != tc.ok {
				t.Errorf("Unexpected result at '%s'", tc.caseName)
			}
		})
	}

}
