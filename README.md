# gistify

Gists are simple way to share small programs and they are even simpler to include in your blog post on Medium or Gitbook. The good thing about inlcuded gist in blog post or gitbook is that they get updated without you need to do much. But the pain comes when you have gist with multiple files and you want to embed individual file from the gist. I think there is no straight forward way or no way at all.

I am planning to write technical blog posts and intent to include gist as necessary. To make my life easier (and yours, if it resonates with you), this is a smalll program, which can create individual gists easily with just a simple command.

Usage:
* Create a 
* Set environment variable GISTIFY_TOKEN with the generated token.
  For example, on MacOS, Linux: `$ export GISTIFY_TOKEN="..."`
* Change to directory where your files are from which gist to be created
* Execute `gistify` command
  For example, `$ gistify`
