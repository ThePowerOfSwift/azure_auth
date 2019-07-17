# Azure oAuth authentication tool

## How to set up application locally

Suppose you already have installed Go and set up $GOPATH

Clone repository in to $GOPATH/src

```bash
git clone https://github.com/Raileanv/azure_auth.git
```

In project directory run `go get ./...`  to install all dependencies.

Next fill up .env file with your credentials. (Use as an example `example.env`)

## How to run application locally

being in project directory run command

```bash
go build
```

This will compile project,
next to start application run 

```bash
./azure_auth
```
(Depending on OS binary might be different)

## URLs

 - [GET] "BASE_URL/auth_url" - Get actual auth url (returns URL to `authentication endpoint`) 
 - [GET] `authentication endpoint` - Use browser for this url (will redirect to Microsoft authentication form
and after all auth steps you will be redirected to some url which contains temporary_token)
 - [POST] "BASE_URL/auth_with_temporary_token?temporary_token=[temporary_token]" (exchange temporary token to public token) 
 - [GET] "BASE_URL/get_me" (in Authorization header put public token) (returns info about user)

