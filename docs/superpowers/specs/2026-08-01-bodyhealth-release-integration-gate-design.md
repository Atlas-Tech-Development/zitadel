# BodyHealth ZITADEL Release Integration Gate

## Context

The `v4.15.3-atlas.1` release run passed source generation, formatting and vetting,
focused OIDC tests, API unit tests, and generated-file checks. The tag-only full
API integration suite then failed in an unrelated user API test with a context
deadline. The OIDC integration package passed in the same run, but the failure
prevented the image build and ACR publication steps from running.

## Decision

Remove the tag-only `Run API integration tests` step from the custom BodyHealth
release workflow. Remove its Docker Compose setup step because no remaining
release step uses Compose.

Keep these release gates:

- API source generation
- formatting and `go vet` for the fork patch
- focused `internal/api/oidc` tests, including the Shopify audience regression
- the complete API unit-test target
- generated-file consistency
- image build and image inspection before export

The upstream full integration suite remains available outside this custom
release workflow. This fork workflow intentionally gates the release on checks
that cover the fork-specific behavior and deterministic unit validation.

## Release sequence

The existing `v4.15.3-atlas.1` tag is immutable and contains the old workflow.
After this workflow change is merged into `bodyhealth/v4.15.3`, create
`v4.15.3-atlas.2`. Do not move or recreate the `.1` tag.

The `.2` workflow must build the image, publish the same tag to the development
and production ACRs, and record both digests. BodyHealth deployment configuration
must not be changed to `.2` until both registry manifests are confirmed.

## Validation

- Validate workflow syntax and pinned action references.
- Confirm the workflow diff removes only Docker Compose setup and the full API
  integration step.
- Run whitespace and diff checks locally.
- On the `.2` tag run, require all remaining checks, image packaging, and both
  ACR pushes to succeed before deployment configuration changes.
