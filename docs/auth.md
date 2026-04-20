# Authentication Setup

Set up Apple Ads API access. `aads` uses Apple Ads OAuth credentials to obtain
and refresh API access tokens automatically.

Quick setup:

```sh
aads profiles create --interactive
```

Or follow the manual steps below.

## Manual Setup

Invite an API user on the [Apple Ads website](https://app-ads.apple.com/cm/app/settings/users)
(skip if you already have a user account with API access):

- Sign in to Apple Ads as an account administrator (or ask an administrator to do these steps).
- Access `User Management` (it should already be open if you used the link above, but make sure it is for the right organization): in the top right, press your name, then press `Settings` -> `User Management`.
- Press `Invite Users`.
- Fill in the fields and choose `API Account Manager` or `API Account Read Only`.

Generate the key pair in a local terminal (skip if you already registered a
private key for API access):

```sh
# Generate a private key and print the public key
aads profiles genkey --name default
```

Set up API access. In a private browser window, or after clearing cookies:

- If a new user was invited, open the invitation link received by email.
- Sign in to Apple Ads with an API-enabled account.
- Access `API Account Settings`: open [this link](https://app-ads.apple.com/cm/app/settings/apicertificates) (make sure it is for the right organization), or in the top right, press your name, then go to `Settings` -> `API`.
- Paste the public key, including delimiters.
- Press `Save`.
- Note the `clientId`, `teamId`, and `keyId` for the next step.

For more details, see [Implementing OAuth for the Apple Ads API](https://developer.apple.com/documentation/apple_ads/implementing-oauth-for-the-apple-search-ads-api).

Create a profile:

```sh
aads profiles create \
    --name default \
    --client-id SEARCHADS.example \
    --team-id SEARCHADS.example \
    --key-id ABC123 \
    --max-daily-budget 1000 \
    --max-bid 10
```

Specify `--default-currency` and `--org-id` if you have access to multiple
currencies or organizations. Otherwise, `aads` uses the first available
currency and org ID.
