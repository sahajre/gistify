Gist is a simple way to share small program. Also, it is very easy to embedd gist in your blog post on Medium or in Gitbook. The good thing about embedded gist in blog post/gitbook is that it gets updated without you need to do anything. But the pain comes when you have gist with multiple files and you want to embed an individual file from it. I don't think there is a way or straight forward way to achieve it.

I am planning to write technical blog posts and intend to include gist as and when necessary. To make my life easier (and yours, if it resonates with you), this is a smalll program, which can create individual gists from set of files easily with just a simple command.

# Usage
* Download `gistify` <br>
`$ go get github.com/sahajre/gistify`
* Create Github access token. 
  * Go to: https://github.com/settings/tokens.
  * click "Generate New Token".
  * select gist and save.
* Set environment variable `GISTIFY_TOKEN` with the generated token.<br>
`$ export GISTIFY_TOKEN="..."`
* Change to target directory
* Execute `gistify` command. <br>
`$ gistify ".*.go"`
* Now, all the matching files will be available as an individual gist and its URL will be shown in the output.
* On subsequent run
  * Unchanged files, will be skipped.
  * For updated file, the gist created earlier will be updated.
  * For a new file, new gist will be created.
