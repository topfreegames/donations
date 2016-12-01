Overview
========

What is Donations? Donations is an HTTP "resty" API for managing donations of items in games.

Donations allows your app to focus on the interaction required to donating items and keeping players engaged, instead of the backend required for actually doing it.

## Features

* **Multi-tenant** - Donations already works for as many games as you need, just keep adding new games;
* **Item Donation Management** - Donate items to other players and request items from them in return;
* **New Relic Support** - Natively support new relic with segments in each API route for easy detection of bottlenecks;
* **Easy to deploy** - Donations comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

## Architecture

Donations is based on the premise that you have a backend server for your game. That means we do not employ any means of authentication.

There's no validation if the actions you are performing are valid as well. We have TONS of validation around the operations themselves being valid.

What we don't have are validations that test whether the source of the request can perform the request (remember the authentication bit?).

## The Stack

For the devs out there, our code is in Go, but more specifically:

* Web Framework - [Echo](https://github.com/labstack/echo) based on the insanely fast [FastHTTP](https://github.com/valyala/fasthttp);
* Database - MongoDB;
* Locks - Redis.

## Who's Using it

Well, right now, only us at TFG Co, are using it, but it would be great to get a community around the project. Hope to hear from you guys soon!

## How To Contribute?

Just the usual: Fork, Hack, Pull Request. Rinse and Repeat. Also don't forget to include tests and docs (we are very fond of both).
