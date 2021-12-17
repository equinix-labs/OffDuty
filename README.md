# OffDuty

Program recurrent PagerDuty overrides based on day of week/time

## Usage

Generate a PagerDuty API token:
 - Log in to PagerDuty
 - In the top right corner hover over the Gravatar and select "My Profile"
 - Select the "User Settings" Tab
 - At the bottom click the "Create API User Token"
 - Save the Token somewhere secure; if you lose it you should delete it and create a new one.
  - If you use direnv and nix, you can place it in a .env file as `PAGERDUTY_TOKEN=yourtoken`

Build offduty

`go build ./cmd/offduty

Run it

`env PAGERDUTY_TOKEN=x offduty --dry-run example.yaml`

## Configuration

See `example.yaml`
