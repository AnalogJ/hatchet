<p align="center">
  <a href="https://github.com/AnalogJ/hatchet">
  <img width="300" alt="hatchet_logo" src="./docs/logo.svg">
  </a>
</p>

# hatchet
Cut down spam in your Gmail Inbox


![](./docs/spreadsheet.png)

# Introduction

Hatchet is a tool to help you manage/prune your Email Inbox.
As it processes your inbox, it will keep track the unique sender email addresses and the number of emails from each sender. 
It will also search the email headers and body for "unsubscribe" links.

Once Hatchet finishes its work, it will generate a spreadsheet that you can use to quickly unsubscribe from the most annoying mailing lists spamming your inbox. 

# Getting Started

## Windows/Mac/Linux Binaries

You can download the latest version of hatchet by visting the following url: https://github.com/analogj/hatchet/releases
Download the relevant binary, then run it:

```bash

# windows 
hatchet-windows-amd64.exe report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx


# macos
hatchet-darwin-amd64 report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx

# linux
hatchet-linux-amd64 report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx
```

## Run from Source

```
go run cmd/hatchet/hatchet.go report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx
```

# Versioning
We use SemVer for versioning. For the versions available, see the tags on this repository.

# Authors
Jason Kulatunga - Initial Development - @AnalogJ

# License

- MIT
- [Logo: Hatchet by Fran Couto from NounProject.com](https://thenounproject.com/icon/hatchet-3263047/)

