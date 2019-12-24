# twtblocks

This script scrapes the Twitter API to find accounts which are blocking the calling user.

As there is no official API endpoint to get this information directly we need to use other means to detect if a user is blocking us.
The [users/lookup](https://developer.twitter.com/en/docs/accounts-and-users/follow-search-get-users/api-reference/get-users-lookup)
method lets us fetch information about many users at once. The user object is of particular interest as it includes a ``Status`` field,
which contains the last Tweet of the user in question.

If the ``Status`` object is nil, whilst the status count (``StatusesCount``) is greater zero*,
and the user's tweets are not protected, we are definitely blocked.
In all other cases with an empty status object we need to call the [friendships/show](https://developer.twitter.com/en/docs/accounts-and-users/follow-search-get-users/api-reference/get-friendships-show) method to find out if we're blocked.
This endpoint is heavily rate limited though, that's why we only use it in those edge cases to achieve a reasonable execution time.

To populate the list of users which should be checked, we use [friends/list](https://developer.twitter.com/en/docs/accounts-and-users/follow-search-get-users/api-reference/get-friends-list) on a few user provided accounts to get their followings.

\* The status count is not always accurate, for example when a user deletes all of his tweets the status count could still be stuck at a value greater than zero, that's why we need the additional check for accounts with a low status count.

## Installation

### From source

```sh
go get -u github.com/buckket/twtblocks
```

## Configuration

- Edit config.toml (Twitter API credentials)

## Usage

```sh
./twtblocks -config config.toml jack elonmusk buckket
```

Fetch all followings of @jack, @elonmusk and @buckket and check if any of those account are blocking the calling user.

## License

GNU GPLv3+
