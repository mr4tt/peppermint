## What's Peppermint?

Peppermint is a budgeting website where you can write your budgeting goals. It will look at your transactions (using Teller) and notify you when you are close to maxing out a certain category (or going over).

## To use Peppermint

We assume you have created a [Teller](https://teller.io/) account.

1. `git clone https://github.com/mr4tt/peppermint`

   - clone Peppermint locally

2. `cd peppermint`

3. Edit `test.html` to include your application id from your Teller dashboard (line 15).

4. Run `test.html` in a browser and connect a bank account with Teller. Your access code will be logged in console.

5. `touch .env`

   - inside .env, put in your Teller Connect access code.
   - `ACCESS_TOKEN=token_xxxxxxxxxxx`

6. `mkdir certs`

   - inside `certs/`, please add your SSL certs from your Teller dashboard.
   - this should be `certificate.pem` and `private_key.pem`

7. `go mod tidy`

   - installs necessary dependencies

8. `go run main.go`
   - this will show transactions in JSON at http://localhost:3000/transactions
