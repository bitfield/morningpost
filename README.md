# The Morning Post

This is a Go project template, intended for students at the [Bitfield Institute of Technology](https://bitfieldconsulting.com/golang/bit)—but everyone's welcome to try it if you're interested in learning or practicing Go.

## Description

It's nice to have something interesting to read with your morning coffee. Your task with this project is to write a Go program that will curate a little “morning newspaper” for you, by scraping news websites and selecting random stories.

For example, it might choose stories from the front page of [Hacker News](https://news.ycombinator.com/news), or [The Guardian](https://www.theguardian.com). Or maybe a selection of different sites: that's up to you!

## Implementation

It's also up to you how those stories are presented to the user. You could simply print out the headlines and links on the terminal, or open them as tabs in the default web browser. Or you could make the Morning Post run as a webserver itself, so that users just browse to some local URL and see the selected stories.

For scraping the news sites, your first choice should be an API. For example, there's a [Hacker News API](https://github.com/HackerNews/API), and the [Guardian OpenPlatform](https://open-platform.theguardian.com/). It's up to you whether you want to use third-party Go packages to talk to these APIs, or write your own.

## Configuration

You might provide some way for users to configure their preferences: for example, which of the supported news sites to get stories from, or how many stories, or perhaps some search keywords that will be used to find stories. For example, you could configure the newspaper to feature stories from Hacker News about “Golang”, if you're a bit of a masochist, or “Elon Musk”, if you've got a lot of coffee to get through.

## Getting started

There are lots of fun possibilities, but I suggest you start by building the simplest possible working implementation that you can. By “working”, I mean that you can run the CLI tool and see some actual news stories, in whatever form, from perhaps just one site to start with.

Once you have a working program, it's easier to think about how to extend it with additional features or news sites. And focusing on the simplest possible implementation to start with can help prevent you getting bogged down in complicated things you might not actually need.

## Going further

Think about how other people might be able to extend your package. For example, if someone wanted to add Reddit as a story source (after all, kids read too), how would they do that?

Would they have to repeat a lot of the existing code, but with some changes to support the new site? That doesn't sound very fun. Better would be if they could implement some kind of interface, and then you could import their package into the Morning Post tool to add support for the new site.

If your project is a big success, you might find many people want to contribute support for their favourite news sites. Every time someone does, you can use their contribution to improve the Morning Post itself. Nice!

## Acknowledgements

Many thanks to [Josh Akeman](https://github.com/joshakeman) for suggesting this project.
