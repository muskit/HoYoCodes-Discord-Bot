## Subscriptions
Subscriptions work on a *per-channel* basis; **one channel gets on subscription**. Each subscription can be configured using the following commands:
- `/subscribe`: Creates a subscription for a channel. This can be run on an already-subscribed channel to reconfigure it with the following options.
  - `announce_code_additions`: Determine if the subscription should notify of new codes being added. Default: `true`
  - `announce_code_removals`: Determine if the subscription should notify of codes being removed. Default: `false`
- `/unsubscribe`: Deactivates a subscription for a channel
- `/filter_games`: Set games that a subscription should notify for. By default, **the subscription will notify for all games**. Specify no games in the command to subscribe to all.
- `/add_ping_role`: Add a role that will be pinged for a channel's subscription.
- `/remove_ping_role`: Remove a role from being pinged for a channel's subscription.