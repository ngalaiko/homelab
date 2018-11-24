---
title: "Half a year with vim"
tags: [
    "vim",
    "development",
]
date: "2018-07-21"
categories: [
    "Blog",
]
---

![Vim learning curve](/media/vim.jpg)

I started using vim because I was bored. Likely, I code in go, so it’s pretty
easy to switch. I mean, I tried to do Java with vim for about a week -
and it’s a hell. So if your primary language requires massive IDE support,
it’s not an option to completely switch to vim.

After slightly more than half a year of usage, I improved typing,
navigation speed, and general comfort while I code.

Before I was using a mouse to navigate between files,
scrolling them to find something, But with vim,
your movements are minimal and efficient.

Also, it makes you think a bit more. Not only because you need to remember a
lot of commands and combinations, but because you always try to do something 
by pressing the minimum number of keys possible.

I am going to describe my experience on how to start and make an
overview of the configuration I use.

## How to start

### Basics

As a first step, to understand the basics, I suggest trying embedded vim tutor. You can run it by:

```
$ vimtutor
```

It should be installed by default.

It’s a text file divided into lessons which will teach you to navigate, edit,
search and more - basics enough for daily use. I was doing this tutorial for a couple of
weeks until I was confident with each lesson.

Once you got used to `hjkl` navigation, stop doing it. The main power of vim is 
navigation and clicking `hhhhhh` to get to another line, for instance, is 
super inefficient. It can be as hard as switching to vim, but you will not 
regret. Here is a [good video](https://www.youtube.com/watch?v=OnUiHLYZgaA) about it.

### Configuration

Some people suggest to use vanilla vim to start with, so you can understand
which functions you luck of comparing to IDE/editor you used to and do not
install a lot of useless plugins.

However, I think it makes sense to find the most popular configuration
for a language you use on GitHub and use it as a base.

After some time you most likely will optimize it for your usage.
I did exactly same, started with [vim-go-ide](https://github.com/farazdagi/vim-go-ide)
and then [forked it](https://github.com/ngalayko/vim-go-ide).

## Configuration overwiew

Here is a list of plugins I use: 

* [gruvbox](https://github.com/morhetz/gruvbox) - really good color scheme (I prefer dark version)
* [deoplete.nvim](https://github.com/Shougo/deoplete.nvim) - fast-auto completion, 
the first thing to have when you used to IDE
* [deoplete-go](https://github.com/zchee/deoplete-go) - go specific completion options.
By default it can complete only words from the same file, recently used words, file paths
* [nvim-yarp](https://github.com/roxma/nvim-yarp) - required by deoplete in order to work with vim8
* [vim-hug-neovim-rpc](https://github.com/roxma/vim-hug-neovim-rpc) - same
* [nerdtree](https://github.com/scrooloose/nerdtree) - the best plugin to navigate 
in a directory tree. Has a lot of options from creating/deleting a file to some super weird I don’t use
* [vim-fugitive](https://github.com/tpope/vim-fugitive) - super powerful git integration
* [nerdcommenter](https://github.com/scrooloose/nerdcommenter) - smart commenting selected lines/parts of code
* [vim-go](https://github.com/fatih/vim-go) - the main plugin if you code in Go. Includes a lot of commands,
syntax checking, syntax highlighting, go tools. Make sure you read documentation
* [auto-pairs](https://github.com/jiangmiao/auto-pairs) - automatically close paired symbols e.x. brackets
* [tagbar](https://github.com/majutsushi/tagbar) - overview of source code. variables / functions / classes / etc.
* [fzf](https://github.com/junegunn/fzf) - if you still use `grep` or `ag`, check this out
* [fzf.vim](https://github.com/junegunn/fzf.vim) - better fzf integration for vim.
* [Dockerfile.vim](https://github.com/ekalinin/Dockerfile.vim) - Dockerfiles syntax highlighting
* [vim-signify](https://github.com/mhinz/vim-signify) - hightlighting of changed lines of code for any vcs
* [nginx.vim](https://github.com/chr4/nginx.vim) - nginx syntax highlighting

# Links 

* [My fork of vim-go-ide](https://github.com/ngalayko/vim-go-ide)
* [How to Do 90% of What Plugins Do (With Just Vim)](https://www.youtube.com/watch?v=XA2WjJbmmoM) 
* [Improving Vim Speed](https://www.youtube.com/watch?v=OnUiHLYZgaA)
