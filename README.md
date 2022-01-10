# hatchet
Cut down spam in your Gmail Inbox


![](./docs/spreadsheet.png)

# Introduction

Hatchet is a tool to help you manage/prune your Email Inbox.
As it processes your inbox, it will keep track the unique sender email addresses and the number of emails from each sender. 
It will also search the email headers and body for "unsubscribe" links.

Once Hatchet finishes its work, it will generate a spreadsheet that you can use to quickly unsubscribe from the most annoying mailing lists spamming your inbox. 

# Getting Started

```

go run cmd/hatchet/hatchet.go report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx

```
