---
title: "Golang: Data races" 
tags: [
    "go",
    "data race",
    "race condition",
]
date: "2019-02-02"
categories: [
    "Blog",
]
---

I have noticed that many people who have started using go have troubles when it
comes to concurrent programming. Concurrency in go is indeed the most 
complicated part of the language, especially for people who don't have much 
experience working with it. There are no compile time validations to prevent a 
programmer from creating race conditions, but go provides all the needed 
tools and instruments to avoid it.

I will try to explain what is a race condition, why does it happen and how to 
avoid it.

Wikipedia:

> A race condition or race hazard is the behavior of an electronics, software,
or another system where the system's substantive behavior is dependent on the
sequence or timing of other uncontrollable events. It becomes a bug when one or
more of the possible behaviors is undesirable.

Let's say we have a list of links to Wikipedia pages that we want to download.

I wrote a [simple tool](https://github.com/ngalayko/examples/tree/master/concurrency/examples/download) as an example that uses the interface to do that:
```go
type Downloader interface {
	Download(...*url.URL) (map[*url.URL][]byte, error)
}
```

The most straightforward implementation would be to iterate over a list of urls,
download each of them and store in the resulting map: 
```go
type Downloader struct {
	client *http.Client
}

func (d *Downloader) Download(urls ...*url.URL) (map[*url.URL][]byte, error) {
	result := make(map[*url.URL][]byte, len(urls))
	for _, u := range urls {
		data, err := d.download(u)
		if err != nil {
			return nil, fmt.Errorf("error downloading %s: %s", u, err)
		}
		result[u] = data
	}
	return result, nil
}

func (d *Downloader) download(u *url.URL) ([]byte, error) {
	resp, err := d.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
```

The code works, but not so fast, because the total time to download links is a sum
of times to download for each link. It's really inefficient when the number of 
links is big. For me, it took over 16s to download 100 links. To improve it, let's 
change the code, so we download all links at the same time:

```go
func (d *Downloader) Download(urls ...*url.URL) (map[*url.URL][]byte, error) {
	result := make(map[*url.URL][]byte, len(urls))

	wg := &sync.WaitGroup{}
	for _, u := range urls {
		wg.Add(1)
		go func(u *url.URL) {
			data, _ := d.download(u)
			result[u] = data
			wg.Done()
		}(u)
	}
	wg.Wait()

	return result, nil
}
```

To do that, we have wrapped our `download` function to run in its goroutine.
Wait group here is used to wait until all of the urls are downloaded, so the 
total time equals the time to download the heavies page. On my test run, it 
is `679.53154ms` for 100 pages. Down from sixteen seconds to less than a second - 
an impressive result! 

But here is a problem:
```
==================
WARNING: DATA RACE
Write at 0x00c0007baf60 by goroutine 257:
  runtime.mapassign_fast64()
      /usr/local/go/src/runtime/map_fast64.go:92 +0x0
  github.com/ngalayko/talks/concurrency/examples/download/downloader/race.(*Downloader).Download.func1()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/downloader/race/downloader.go:37 +0x8e

Previous write at 0x00c0007baf60 by goroutine 262:
  runtime.mapassign_fast64()
      /usr/local/go/src/runtime/map_fast64.go:92 +0x0
  github.com/ngalayko/talks/concurrency/examples/download/downloader/race.(*Downloader).Download.func1()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/downloader/race/downloader.go:37 +0x8e

Goroutine 257 (running) created at:
  github.com/ngalayko/talks/concurrency/examples/download/downloader/race.(*Downloader).Download()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/downloader/race/downloader.go:35 +0x123
  main.run()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/cmd/download/main.go:71 +0x8e
  main.main()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/cmd/download/main.go:56 +0x3b4

Goroutine 262 (finished) created at:
  github.com/ngalayko/talks/concurrency/examples/download/downloader/race.(*Downloader).Download()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/downloader/race/downloader.go:35 +0x123
  main.run()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/cmd/download/main.go:71 +0x8e
  main.main()
      /Users/nikitagalaiko/golang/src/github.com/ngalayko/talks/concurrency/examples/download/cmd/download/main.go:56 +0x3b4
==================
```

I've built the code with `-race` flag to enable golang race detector and got a 
warning about the data race. Of cource, a warning is not an error, and we can 
ignore it, the result was correct: we've downloaded a page for each link. 

But why race detector does not agree?

First, let's see what lines he doesn't like: 
```go 
    result[u] = data
```

A map is a data structure that contains keys and values. All keys are unique, 
and each key has an associated value with it. For the sake of the example 
let's say that every key is associated with a particular address in the memory, 
and the address is permanent. 

So when we are updating the key's value, we are updating data at the same address 
in the memory. 

In our example, we have unique urls, we download each of them and place the content 
to the individual memory address associated with a map key.

But what should happen when we write different data into the same key? If we 
follow the abstractions of a map that I've described above and goroutines - 
light-weight threads that are running concurrently at the same time - then the 
final value for the map is not determined. 

Because what happens is that we are telling go to write some data at the same 
memory address multiple times at the same time. This can cause all kinds of 
problems - from the unexpected result for a programmer to data corruption. And it's 
not a responsibility of a language to decide what does a programmer wants to see 
as a result. And a result in the situation depends on uncontrollable events.

That's why go gives us a warning. It can't guarantee the result.

It doesn't apply only to maps. The same can happen when you share a pointer 
between two goroutines. Or you are writing to the same file from two different 
programmes, or maybe you are saving data from two instances of your applications to 
the same database. These all are different kinds of the same problem. 

> Data races happen only when you have multiple threads/goroutines/processes that 
can access the same data at the same time. 

Likely, that problem was around for a while, and there is a solution to it. 

> You need to make sure that only one  process has access to the same piece of
memory at a time.

Though, there is a difference between write and read access. Reading the same
data from multiple threads is safe. The problems start only when you have
a thread that writes.

Golang has a few instruments to deal with data races like mutexes and channels.

Mutexes are used to solve exactly data race problems, here how to use them: 
```go
func (d *Downloader) Download(urls ...*url.URL) (map[*url.URL][]byte, error) {
	result := make(map[*url.URL][]byte, len(urls))
	wg := &sync.WaitGroup{}

	guard := &sync.Mutex{}

	for _, u := range urls {
		wg.Add(1)
		go func(u *url.URL) {
			data, _ := d.download(u)

			guard.Lock()
			result[u] = data
			guard.Unlock()

			wg.Done()
		}(u)
	}
	wg.Wait()
	return result, nil
}
```
It has two methods, `Lock` and `Unlock`. Everything between them can be accessed
only by one goroutine at a time. Others will wait for the mutex to unlock before 
they can access it.

Channels are a bit more tricky, and can be also used to controll goroutines:
```go
type done struct {
	u    *url.URL
	data []byte
	err  error
}

func (d *Downloader) Download(urls ...*url.URL) (map[*url.URL][]byte, error) {
	result := make(map[*url.URL][]byte, len(urls))

	doneChan := make(chan *done)

	wg := &sync.WaitGroup{}
	for _, u := range urls {
		wg.Add(1)
		go func(u *url.URL) {
			data, err := d.download(u)
			doneChan <- &done{
				u:    u,
				data: data,
				err:  err,
			}
			wg.Done()
		}(u)
	}

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	for done := range doneChan {
		if done.err != nil {
			return nil, fmt.Errorf("error downloading %s: %s", done.u, done.err)
		}
		result[done.u] = done.data
	}

	return result, nil
}
```

Data from the multiple goroutines sent to the single goroutine via a channel and 
that goroutine stores the data in the map.

## Links: 
* [example](https://github.com/ngalayko/examples/tree/master/concurrency/examples/download)
