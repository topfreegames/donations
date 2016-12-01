Game Configuration
==================

Being a multi-tenant server, Donations allows for many different configurations per tenant. Each tenant is a different game and is identified by it's game ID.

Before any operation can be performed, you must create a game in Donations. The good news here is that updating games are idempotent operations. You can keep executing it any time your game changes. That's ideal to be executed in a deploy script, for instance.

## Creating/Updating a Game

The `Update` operation of the Game is idempotent (you can run it as many times as you want with the same result). If your game does not exist yet, it will create it, otherwise just updated it with the new configurations.

To Create/Update your game, just do a `PUT` request to `http://my-donations-server/games/my-game-public-id`, where `my-game-public-id` is the ID you'll be using for all your game's operations in the future. The payload for the request is a JSON object in the body and should be as follows:

```
    {
      "name":                          [string],
      "donationCooldownHours":         [int],
      "donationRequestCooldownHours":  [int]
    }
```

If the operation is successful, you'll receive a JSON object with the game details.

## Game Configuration Settings

As can be seen from the previous section, there are some configurations you can do per game. These will be thoroughly explained in this section.

### name

The name of your game. This is used mainly for easier reasoning of what this game is when debugging.

**Type**: `string`<br />
**Sample Value**: `My Sample Game`

### donationCooldownHours

Number of hours that must ellapse before a player can donate again. This is used in conjunction with the max donation weight passed in with the donation.

**Type**: `Integer`<br />
**Sample Value**: `24`

### donationRequestCooldownHours

Number of hours a player must wait before requesting donations again.

**Type**: `Int`<br />
**Sample Value**: `8`

## Game Items

In order to use donations, the items that can be donated must be previously created for that specific game.

As with the game, creating/updating items can be done idempotently, thus allowing for easier maintenance of a game's donatable items.

### Creating/Updating an Item

The `Update` operation of the Item is idempotent (you can run it as many times as you want with the same result). If your item does not exist yet, it will create it, otherwise just updated it with the new configurations.

To Create/Update your item, just do a `PUT` request to `http://my-donations-server/games/my-game-public-id/items/my-item-key`, where `my-game-public-id` is the ID of the game you registered in the previous step and `my-item-key` is the item key you'll be using in your game's donation requests and donations in the future. The payload for the request is a JSON object in the body and should be as follows:

```
    {
      "metadata":                           [JSON],
      "weightPerDonation":                  [int],
      "limitOfItemsPerPlayerDonation":      [int],
      "limitOfItemsInEachDonationRequest":  [int]
    }
```

If the operation is successful, you'll receive a JSON object with the item details.

### Item Configuration Settings

As can be seen from the previous section, there are some configurations you can do per item. These will be thoroughly explained in this section.

### metadata

This sections allow you to store metadata for this item.

Donations treats this as a black box and won't use it for any operations. It will be returned whenever the item is returned, though.

**Type**: `JSON`<br />
**Sample Value**: `{ "cost": 100, "currency": "gold" }`

### weightPerDonation

Donations keep track of the weight that has been donated by player in case you want to limit the max weight donated per player per time.

This configuration allows you to set different weights per donation per item (epic donations weigh more than rare and so on).

**Type**: `int`<br />
**Sample Value**: `3`

### limitOfItemsPerPlayerDonation

The limit of items that can be donated by a single player for this item in a single donation request.

**Type**: `int`<br />
**Sample Value**: `3`

### limitOfItemsInEachDonationRequest

The amount of this item that can be donated in a single donation request.

**Type**: `int`<br />
**Sample Value**: `8`
