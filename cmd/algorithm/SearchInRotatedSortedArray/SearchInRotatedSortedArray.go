package main

import "github.com/go/src/fmt"

func main() {
	arr := []int{4,5,6,7,0,1,2}
	SearchInRotatedSortedArrayI(arr)
}

func SearchInRotatedSortedArrayI(input []int) {
	fmt.Println(findMin1(input))
	fmt.Println(findMin2(input))
}

func SearchInRotatedSortedArrayII() {

}

func findMin1(input []int) int {
	var i = int(1)
	for ; i < len(input); i++ {
		if input[i] < input[i-1] {
			break
		}
	}
	if i == len(input) {
		return input[0]
	}
	return i
}

func findMin2(nums []int) int {
	left := int(0)
	right := len(nums) - 1
	for ; left < right; {

		if nums[right] > nums[left] {
			return nums[left]
		}

		mid := int(left + (right-left)/2)

		if nums[mid] >= nums[left] {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return nums[right]
}
