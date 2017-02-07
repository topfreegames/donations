Donations API
=============

## Healthcheck Routes

  ### Healthcheck

  `GET /healthcheck`

  Validates that the app is still up, including the database connection.

  * Success Response
    * Code: `200`
    * Content:

      ```
        "WORKING"
      ```

  * Error Response

    It will return an error if it failed to connect to the database.

    * Code: `500`
    * Content:

      ```
        "Error connecting to database: <error-details>"
      ```

## Game Routes

  ### Update Game
  `PUT /games/:gameID`

  Updates the game with that has publicID `gameID`.

  * Payload

    ```
    {
      "name":                          [string],  // 2000 characters max
      "donationCooldownHours":         [int],
      "donationRequestCooldownHours":  [int]
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "name":                          [string],  // 2000 characters max
        "donationCooldownHours":         [int],
        "donationRequestCooldownHours":  [int]
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Item Routes

  ### Update Item
  `PUT /games/:gameID/items/:itemKey`

  Updates the item with key `itemKey` in the game with public ID `gameID`.

  * Payload

    ```
    {
      "metadata":                           [JSON],
      "weightPerDonation":                  [int],
      "limitOfItemsPerPlayerDonation":      [int],
      "limitOfItemsInEachDonationRequest":  [int]
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "metadata":                           [JSON],
        "weightPerDonation":                  [int],
        "limitOfItemsPerPlayerDonation":      [int],
        "limitOfItemsInEachDonationRequest":  [int]
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Donation Request Routes

  ### Create Donation Request
  `POST /games/:gameID/donation-requests`

  Creates a new donation request for the item specified in the payload.

  * Payload

    ```
    {
      "item":      [string],
      "player":    [string],
      "clan":      [string],
    }
    ```

    * `item` is the key for the item to create the donation request for;
    * `player` is the player id that will receive the donations;
    * `clan` is the team/clan/group the player belongs to. This is useful for grouping donations. Leave this empty if player does not belong to a team/clan/group.

  * Success Response
    * Code: `200`
    * Content: Serialized donation request.

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Donation Routes

  ### Donate to a Donation Request
  `POST /games/:gameID/donation-requests/:donationRequestID`

  Donates items to a donation request with public ID `donationRequestID` in the game `gameID`.

  * Payload

    ```
    {
      "player":                 [string],
      "amount":                 [string],
      "maxWeightPerPlayer":     [string]
    }
    ```

    * `player` is the player id that will receive the donations;
    * `amount` is the quantity of the item being donated by this player;
    * `maxWeightPerPlayer` is the maximum weight this player can donate per time period.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```
