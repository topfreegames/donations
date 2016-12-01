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
