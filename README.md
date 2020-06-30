# Go-Safe
Go Webapp - Password Manager

Go-Safe is a password manager based on AES, it is totally safe beacause the badge (used for making the key) is not saved on database, so not even the Database's admin can read your passwords. You can find Go-Safe [here](https://go-safe.herokuapp.com/home)

## How is the key generated?

The formula for generating the key is simple and complex at the same time:

```
SHA256(SHA256(badge) + SHA256(password))
```
So an hacker must stealing/guess your password and your badge for accesing your secret area. If a hacker were to take the control of our database still he could not read your passwords. (he has not the badge)



