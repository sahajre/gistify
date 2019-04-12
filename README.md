# gistify

Gists are simple way to share small programs and they are even simpler to include in your blog post on Medium or Gitbook. The good thing about inlcuded gist in blog post/gitbook is that it gets updated without you need to do anything. But the pain comes when you have gist with multiple files and you want to embed an individual file from the multi-file gist. I think there is no straight forward way or no way at all.

I am planning to write technical blog posts and intend to include gist as and when necessary. To make my life easier (and yours, if it resonates with you), this is a smalll program, which can create individual gists easily with just a simple command.

Usage:
* Download `gistify` <br>
`$ go get github.com/sahajre/gistify`
* Create Github access token. <br>
Go to: https://github.com/settings/tokens <br>
click "Generate New Token" <br>
select gist and save <br>
* Set environment variable `GISTIFY_TOKEN` with the generated token.<br>
`$ export GISTIFY_TOKEN="..."`<br>
* Change to target directory
* Execute `gistify` command. <br>
`$ gistify ".*.go"`
* Now, all the matching files will be available as an individual gist and the gist URL will be shown in the output.
* On subsequent run, the files which are not updated will be skipped. For updated files, the gist created earlier will get updated. For new files, new gist will be created.
