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
 - [GET] "BASE_URL/get_user_photo" (in Authorization header put public token) (returns blob)
 
_also you can use postman collection_ `azureGoAuth.postman_collection.json`

## How to deploy on heroku 

[info here](https://www.google.com/search?source=hp&ei=BRkwXdPOCYqOrwTc35uwDA&q=heroku+deploy+golang+app&oq=golang+deploy+heroku&gs_l=psy-ab.1.1.0j0i8i30.2327.15465..18521...0.0..0.145.2726.0j21......0....1..gws-wiz.....0..0i10j0i7i30j0i8i7i30.7RQ-mSRSH5c)