---
title: "Daily Coding Problem #23"
tags: [
    "go",
    "development",
    "interview",
    "maze",
]
date: "2018-09-01"
categories: [
    "Daily Coding Problem",
]
---

![maze](/media/maze.jpg)

# Problem 

You are given an M by N matrix consisting of booleans that represents a board.
Each True boolean represents a wall. Each False boolean represents a tile you can walk on.

Given this matrix, a start coordinate, and an end coordinate, return the 
minimum number of steps required to reach the end coordinate from the start.
If there is no possible path, then return null. You can move up, left, down,
and right. You cannot move through walls. You cannot wrap around the edges of the board.

For example, given the following board:
```
[
    [f, f, f, f],
    [t, t, f, t],
    [f, f, f, f],
    [f, f, f, f],
]
```
and start = (3, 0) (bottom left) and end = (0, 0) (top left), the minimum number
of steps required to reach the end is 7, since we would need to go through,
(1, 2) because there is a wall everywhere else on the second row.


# Solution

It is a new type of problems I faced. I remember, I solved some during
university, but it was pretty hard to come up with the solution right away. 

I googled basic types of maze solving algorithms, and it looks like 
[Lee algorithm](https://en.wikipedia.org/wiki/Lee_algorithm) will be a pretty
good choice in most of the `shortest path` problems since at the end of the
day a number of different paths in a maze makes a tree.

The idea is deadly simple: 

1. go to start cell, mark it `0`.
2. mark all neighbors as `+1`. It is a distance to the starting cell
3. make the same for each of the neighbors

By running this algorithm for each cell, we will get the number of steps it
takes to get to any other point from the start. Of course, we should ignore
walls and previously marked cells on each iteration. 

This is a basic solution and can be optimized for a given problem. 
For example, we can stop our recursive calls once we meet finish cell.

# Code
```go
// the maze is a matrix that represents a maze.
// all cells have value 0, and all walls have value 1.
// start and finish are arrays of 2 elements, [i,j] of the cells.
func solution(maze [][]int, start []int, finish []int) int {                                                                                                               
        // mark start cell is -1
        maze[start[0]][start[1]] = -1

        // mark all cells starting from start recursively
        mark(maze, start, 0)

        // return value of a finish cell
        return maze[finish[0]][finish[1]]
}

// pos is a structure to hold cell coordinates, 
// because []int can't ba used as a map key ¯\_(ツ)_/¯
type pos struct {
        i int
        j int
}

// mark marks all neighbors of a given cell with n+1
func mark(maze [][]int, point []int, n int) {
        i, j := point[0], point[1]

        neighbors := map[pos]bool{
                pos{i + 1, j}: false,
                pos{i - 1, j}: false,
                pos{i, j - 1}: false,
                pos{i, j + 1}: false,
        }
        for p := range neighbors {
                neighbors[p] = markP(maze, p.i, p.j, n+1)
        }

        for p, ok := range neighbors {
                if ok {
                        mark(maze, []int{p.i, p.j}, n+1)
                }
        }
}

// markP used to mark maze[i][j] with given n if exists and not marked.
// returns true if it was marked, otherwise false.
func markP(maze [][]int, i, j, n int) bool {
        if i >= len(maze) || j >= len(maze[0]) {
                return false
        }
        if i < 0 || j < 0 {
                return false
        }
        if maze[i][j] != 0 {
                return false
        }

        maze[i][j] = n
        return true
}
```

# Links

[github](https://github.com/ngalayko/dcp/tree/master/problems/2018-09-01)
