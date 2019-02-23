package main

import (
	"fmt"
	"go4.org/sort"
)

func SubSeqOfNSumEqualM(input []int, M int) [][]int {
	ret := [][]int{}
	fmt.Println(input)
	sort.Ints(input)
	return ret
}

func main() {
	input := []int{1, 2, 3, 4, 5, 6, 7, 1, 2}
	fmt.Println(SubSeqOfNSumEqualM(input, 7))
}

/**初步思路
  问题分析：
     若在序列都是正数的情况下，
     若求和的目标，大于数组中最大元素，则肯定无解，
     因此，目标的解必定存在与小于目标的元素中，我们从序列中，取出一个元素后v1，
     计算剩余的值为：  (target -  v1);
     此时，问题  从剩余的序列中 s1' 中获取  目标值为 (target-v1) 的解。
     若我们一次递归，则可以求解中最终结果；
 1. 先对该序列排序，
 2. 找出当前待计算的数据，从最小的一个数据看，



 */
