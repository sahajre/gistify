# gistify

Gists are simple way to share small programs and they are even simpler to include in your blog post on Medium or Gitbook. The good thing about inlcuded gist in blog post/gitbook is that it gets updated without you need to do anything. But the pain comes when you have gist with multiple files and you want to embed an individual file from the multi-file gist. I think there is no straight forward way or no way at all.

I am planning to write technical blog posts and intend to include gist as and when necessary. To make my life easier (and yours, if it resonates with you), this is a smalll program, which can create individual gists easily with just a simple command.

Usage:
* Create Github access token. Go to: https://github.com/settings/tokens, click "Generate New Token", select gist
* Set environment variable GISTIFY_TOKEN with the generated token.
  For example, on MacOS, Linux: `$ export GISTIFY_TOKEN="..."`
* Change to target directory
* Execute `gistify` command.
  For example, `$ gistify ".*.go"`
