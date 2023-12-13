<p align="center">
  <a href="https://github.com/AnalogJ/hatchet">
  <img width="300" alt="hatchet_logo" src="./docs/logo.svg">
  </a>
</p>

# hatchet

Cut down spam in your Gmail Inbox

![Screen shot of the resultent spreadsheet report.](./docs/spreadsheet.png)

# Introduction

Hatchet is a tool to help you manage/prune your Email Inbox.
As it processes your inbox, it will keep track the unique sender email addresses and the number of emails from each sender.
It will also search the email headers and body for "unsubscribe" links.

Once Hatchet finishes its work, it will generate a spreadsheet that you can use to quickly unsubscribe from the most annoying mailing lists spamming your inbox.

# Getting Started

## Windows/Mac/Linux Binaries

You can download the latest version of hatchet by visting the following url: <https://github.com/analogj/hatchet/releases>
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

## Run via Docker

```bash
docker run --rm -v `pwd`:/out/ \
    ghcr.io/analogj/hatchet:latest report \
    --output-path="/out/sender_report.csv" \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx
```

## Run from Source

```bash
go run cmd/hatchet/hatchet.go report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=xxxxxxx
```

# Google Account Authentication & App Passwords

> If an app or site doesn’t meet Google's security standards, Google might block anyone who’s trying to sign in to your
> account from it. Less secure apps can make it easier for hackers to get in to your account, so blocking sign-ins from
> these apps helps keep your account safe.
>
> [Less secure apps & your Google Account](https://support.google.com/accounts/answer/6010255?hl=en#zippy=%2Cif-less-secure-app-access-is-on-for-your-account%2Cif-less-secure-app-access-is-off-for-your-account%2Cuse-an-app-password)

By default Google will block third party applications from accessing your Gmail account via username + password.
To use `hatchet` with your Gmail account, you'll need to authenticate to your account by doing one of the following

## Option 1: Enable "Less secure app" access

- Go to the [Less secure app access](https://myaccount.google.com/lesssecureapps) section of your Google Account. You might need to sign in.
- Turn **Allow less secure apps** on.

## Option 2: Create an App Password (required for 2FA protected accounts)

If you use [2-Step-Verification](https://support.google.com/accounts/answer/185839) and get a "password incorrect" error when you sign in, you can try to use an App Password.

> An App Password is a 16-digit passcode that gives a less secure app or device permission to access your Google Account. App Passwords can only be used with accounts that have 2-Step Verification turned on.
>
> <https://support.google.com/accounts/answer/185833>

- Go to your [Google Account Settings Page](https://myaccount.google.com/).
- Select Security.
- Under "How you sign in to Google," select **2-Step Verification**. You may need to sign in. If you don’t have this option, it might be because:
  - 2-Step Verification is not set up for your account.
  - 2-Step Verification is only set up for security keys.
  - Your account is through work, school, or other organization.
  - You turned on Advanced Protection.
- At the bottom of the next page, choose **App Passwords** in the **App Name** field enter "hatchet" then click **Create**.
- A new box will appear titled **Generated app pasword**, below will be a passcode (eg. `qwer tyui opas dfgh`). Store it securely in a password manager or similar.
- Tap Done.

## Use your credentials with hatchet

Now that you have the correct credentials to authenticate to your Gmail account with Hatchet, you can run the tool locally.

```bash
hatchet report \
    --imap-hostname=imap.gmail.com \
    --imap-username=xxxxx@gmail.com \
    --imap-password=[gmail account password OR app password]
```

# Versioning

We use SemVer for versioning. For the versions available, see the tags on this repository.

# Authors

Jason Kulatunga - Initial Development - @AnalogJ

# License

- MIT
- [Logo: Hatchet by Fran Couto from NounProject.com](https://thenounproject.com/icon/hatchet-3263047/)
