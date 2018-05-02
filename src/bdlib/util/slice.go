package util

import (
	"math/rand"
	"time"
)

type reducetype func(interface{}) interface{}
type filtertype func(interface{}) bool

// InSlice item是否存在于slice中
func InSlice(item string, slice []string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// InIntSlice int 型 是否在 slice 中.
func InIntSlice(v int, sl []int) (ok bool) {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}

	return false
}

// InInt64Slice int64 型 是否在 slice 中.
func InInt64Slice(v int64, sl []int64) (ok bool) {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}

	return false
}

// InInterfaceSlice 判断 interface 类型的是否在 slice 中.
func InInterfaceSlice(v interface{}, sl []interface{}) (in bool) {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}

	return false
}

// SliceRandList slice 随机打乱顺序.
func SliceRandList(min, max int) (sl []int) {
	if max < min {
		min, max = max, min
	}
	length := max - min + 1
	t0 := time.Now()
	rand.Seed(int64(t0.Nanosecond()))
	list := rand.Perm(length)
	for index := range list {
		list[index] += min
	}

	return list
}

// SliceMerge 合并 slice.
func SliceMerge(slice1, slice2 []interface{}) (sl []interface{}) {
	sl = append(slice1, slice2...)

	return
}

// SliceReduce 经过 reduce 函数后重新生成一个 slice.
func SliceReduce(slice []interface{}, fc reducetype) (dslice []interface{}) {
	for _, v := range slice {
		dslice = append(dslice, fc(v))
	}

	return
}

// SliceRand 随机取出 slice 中的一个元素.
func SliceRand(a []interface{}) (b interface{}) {
	randnum := rand.Intn(len(a))
	b = a[randnum]

	return
}

// SliceSum 计算 slice 加和.
func SliceSum(intslice []int64) (sum int64) {
	for _, v := range intslice {
		sum += v
	}

	return

}

// SliceFilter slice 过滤函数.
func SliceFilter(slice []interface{}, a filtertype) (ftslice []interface{}) {
	for _, v := range slice {
		if a(v) {
			ftslice = append(ftslice, v)
		}
	}

	return
}

// SliceDiff 取出 slice1 中不在 slice2 中的数组元素, 即 slice1 和 slice2 的区别.
func SliceDiff(slice1, slice2 []interface{}) (diffslice []interface{}) {
	for _, v := range slice1 {
		if !InInterfaceSlice(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}

	return
}

// SliceChunk ...
func SliceChunk(slice []interface{}, size int) (chunkslice [][]interface{}) {
	if size >= len(slice) {
		chunkslice = append(chunkslice, slice)
		return
	}
	end := size
	for i := 0; i <= (len(slice) - size); i += size {
		chunkslice = append(chunkslice, slice[i:end])
		end += size
	}

	return
}

// SliceRange 生成一个 int range 的 slice.
func SliceRange(start, end, step int64) (intslice []int64) {
	for i := start; i <= end; i += step {
		intslice = append(intslice, i)
	}

	return
}

// SlicePad 为 slice 添加 size 长度的 val .
func SlicePad(slice []interface{}, size int, val interface{}) (rs []interface{}) {
	if size <= len(slice) {
		return slice
	}
	for i := 0; i < (size - len(slice)); i++ {
		slice = append(slice, val)
	}

	return slice
}

// SliceUnique slice 去重.
func SliceUnique(slice []interface{}) (uniqueslice []interface{}) {
	for _, v := range slice {
		if !InInterfaceSlice(v, uniqueslice) {
			uniqueslice = append(uniqueslice, v)
		}
	}

	return
}

// SliceShuffle slice 随机打乱.
func SliceShuffle(slice []interface{}) (rs []interface{}) {
	length := len(slice)
	for i := 0; i < length; i++ {
		a := rand.Intn(length)
		b := rand.Intn(length)
		slice[a], slice[b] = slice[b], slice[a]
	}

	return slice
}
