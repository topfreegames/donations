Donations Errors
================

All errors include a code as well as a description. Errors in the Donations API conform to a `SerializableError` interface with a `ToJSON` method. Whenever an error occurs in any of the routes, this is the payload that will be returned.

An example:

```
    DonationCooldownViolatedError returns {
      'code': 'DO-001',
      'gameId': 'some-game',
      'playerId': '701E8858-B3F1-421C-805D-5B2EC01B335A',
      'totalWeightForPeriod': 30,
      'maxWeightForPeriod': 20
    }
```

## Game Errors


