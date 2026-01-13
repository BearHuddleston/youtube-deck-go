# Security Model

## Current Design

YouTube Deck is designed as a **single-user application**. It does not implement
user authentication or authorization for data endpoints.

## Deployment Recommendations

When deploying YouTube Deck:

1. **Local use only**: Run on localhost for personal use
2. **Network protection**: If exposing to a network, use a reverse proxy with
   authentication (e.g., nginx with basic auth, Authelia, Authentik)
3. **Do not expose to the public internet** without additional authentication

## Endpoints Without Authorization

The following endpoints modify data without authentication:
- `POST /subscriptions` - Add subscription
- `DELETE /subscriptions/{id}` - Delete subscription
- `POST /subscriptions/{id}/refresh` - Refresh videos
- `PATCH /subscriptions/{id}/active` - Toggle active status
- `PATCH /subscriptions/{id}/hide-shorts` - Toggle shorts filter
- `POST /subscriptions/reorder` - Reorder subscriptions
- `POST /videos/{id}/watched` - Mark video watched

## OAuth Authentication

OAuth with Google is optionally available for importing YouTube subscriptions.
This does NOT protect the data endpoints above.
