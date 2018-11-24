---
title: "Daily Coding Problem #15"
tags: [
    "go",
    "development",
    "interview",
]
date: "2018-07-16"
categories: [
    "Daily Coding Problem",
]
---

Today problem is a `probability` problem.

# Problem 

This problem was asked by Facebook.

Given a stream of elements too large to store in memory, pick a random element from the stream with uniform probability.

# Solution

There are many variations of such problems, and before solving it, I want to show some basic examples that I met.

Most trivial one is **picking one random element from an array**.

Every programming language has a function to generate a pseudo-random number (`int` or `float`) within the given range. If you think
of an array **A** as of **N** numbers, it's clear how to pick random one: generate number **i** from 0 to N, and **a[i]** is the answer. 
```go
func oneRandom(a []int) int {
    i := rand.Intn(len(a))
    return aa[i]
}
```

Sometimes you need to pick **N random elements from an array**. 

In this case, you can use the same approach and pick **N** indexes from **0** to **len(A)**.
```go
func nRandom(a []int, n int) []int {
    result := make([]int, n) 
    for i := 0; i < n; i++ {
        result[i] = oneRandom(a)
    }
    return result
}
```

However, if you need to pick **N** random elements, they also have to be **different**.

In this case, you can pick **N** different indexes and make sure that they are different:
```go
func nDifferentRandom1(a []int, n int) []int {
    m := map[int]bool{}
    for len(m) != n {
        j := rand.Intn(len(a))
        m[j] = true
    }

    result := make([]int, n)
    for i := range m {
        result[i] = a[i]
    }
    return result
}
```

But it’s not very efficient, because you can spend much time trying to pick the
last index when you need to pick 8 elements from an array of length 10.

In this case, you can use the well-known zero-allocation algorithm to do so.
The idea is to move elements you picked to the beginning of an array,
and choose from others next:

```go
func nDifferentRandom2(a []int, n int) []int{
    for i := 0; i < n; i++ {
        chooseFrom := a[i:]                            // define a slice to pick from
        choosenIndex := rand.Intn(len(chooseFrom))     // pick a random index from i to N
        a[i], a[choosenIndex] = a[choosenIndex], a[i]  // swap choosenElement with i
    }
    // after all, we have choosen elements at the 
    // begining of an array, and probability is always same.
    return a[:n] 
}
```

Let's return to problem #15 and try to solve it using previous examples.

What do we have:

1. a stream of elements too large to store in memory
2. should pick 1 random element

We understand that to pick a random element from the stream with uniform probability,
we need to process hole stream somehow (without storing to memory).

Well, we always can iterate over a stream (but just once). 

If we knew the length of a stream, we could use the approach from the first example, that would be perfect.

In first example each element of an array have a linked number - it's index. 
Moreover, we used a function that returns number from 0 to n with uniform probability within that range to pick one.

So, in this case, we can **assign** a random value to each element with the same probability, and choose between them based on it linked number.

Let's do it: for each element, we generate a float number between **[0..1]**. Also, we remember 
maximum value that we generated, an element from the stream associated with it - random element. 

# Code
```go
func solution(in <-chan int) <-chan int {
	res := make(chan int, 1)
	go func() {
		var result int
		var lastprob float64
		for a := range in {
			prob := rand.Float64()
			if prob > lastprob {
				result = a
				lastprob = prob
			}
		}
		res <- result
	}()
	return res
}
```

# Links

[github](https://github.com/ngalayko/dcp/tree/master/problems/2018-07-16)
