# Changelog

## v0.5.4 - 2024-08-23

### Bug Fixes

- specs: handle deletion when reading version @joshbeard (#156)

## v0.5.3 - 2024-08-23

### Bug Fixes

- fix: delete missing docs @joshbeard (#155)

## v0.5.2 - 2024-08-23

### Changes

- docs: minor touchups @joshbeard (#154)

## v0.5.0 - 2024-08-23

### Features

- fix: deleting nested docs @joshbeard (#153)
  - A new `config` parameter is added to the provider's configuration that contains a single `destroy_child_docs` attribute now that can toggle the behavior of the provider when deleting a doc with children. Previously, the provider would simply fail. With this change, a user can enable the provider to delete nested docs as they're encountered or fail with more helpful output. This fix was to address certain edge cases with managing docs for an API reference that had implicit child docs.
  - Deleting a doc that doesn't exist (slug not found) will now remove the resource from state and emit a warning. Previously, the provider would throw an error and the only recourse was to manually remove the resource from state.
  - Update the behavior of `use_slug` - previously, the provider would mark the resource for re-creation if `use_slug` as modified. This wasn't necessary and could lead to unintended side-effects. The provider will now remove the resource from state if the doc is not found remotely and emit a warning.
  

### Changes

- API Specifications: refactors for clarity; doc improvements @joshbeard (#152)
- Update readme-api-client: pagination @joshbeard (#151)
- build(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.10.0 to 1.11.0 @dependabot (#149)

## v0.4.0 - 2024-08-02

### Features

- Update a doc's slug using frontmatter @joshbeard (#147)

### Bug Fixes

- Doc slug fixes; API specs data source versions fix @joshbeard (#147)

### Maintenance

- Dependency Updates - Improves error output via readme-api-go-client update @joshbeard (#148)
- build(deps): bump github.com/liveoaklabs/readme-api-go-client from 0.2.2 to 0.2.3 @dependabot (#143)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.18.0 to 0.19.4 @dependabot (#141)
- build(deps): bump goreleaser/goreleaser-action from 5.0.0 to 6.0.0 @dependabot (#139)
- build(deps): bump golang.org/x/vuln from 1.0.4 to 1.1.2 @dependabot (#142)
- build(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.7.0 to 1.9.0 @dependabot (#140)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.33.0 to 2.34.0 @dependabot (#136)
- build(deps): bump securego/gosec from 2.19.0 to 2.20.0 @dependabot (#135)
- build(deps): bump github.com/hashicorp/terraform-plugin-go from 0.22.1 to 0.23.0 @dependabot (#133)
- build(deps): bump actions/setup-go from 5.0.0 to 5.0.1 @dependabot (#131)

## v0.3.3 - 2024-04-04

### Bug Fixes

- Gracefully handle deleted docs and changelogs @joshbeard (#125)

### Maintenance

- build(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.6.1 to 1.7.0 @dependabot (#124)
- Dependency updates @joshbeard (#123)

## v0.3.2 - 2024-02-15

### Changes

- Bump API client; omit empty frontmatter @joshbeard (#116)

### Bug Fixes

- fix: changelog title validation @joshbeard (#117)

## v0.3.1 - 2024-02-14

### Bug Fixes

- fix: changelog 'type' is optional @joshbeard (#115)

### Maintenance

- build(deps): bump securego/gosec from 2.18.2 to 2.19.0 @dependabot (#114)
- build(deps): bump golang.org/x/vuln from 1.0.3 to 1.0.4 @dependabot (#113)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.31.0 to 2.32.0 @dependabot (#112)
- build(deps): bump github.com/segmentio/golines from 0.11.0 to 0.12.2 @dependabot (#110)
- build(deps): bump release-drafter/release-drafter from 5 to 6 @dependabot (#111)
- build(deps): bump mvdan.cc/gofumpt from 0.5.0 to 0.6.0 @dependabot (#108)
- build(deps): bump github.com/hashicorp/terraform-plugin-go from 0.20.0 to 0.21.0 @dependabot (#107)
- build(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.16.0 to 0.18.0 @dependabot (#106)
- build(deps): bump golang.org/x/vuln from 1.0.1 to 1.0.3 @dependabot (#109)
- build(deps): bump github.com/liveoaklabs/readme-api-go-client from 0.1.3 to 0.2.0 @dependabot (#102)

## v0.3.0 - 2024-01-19

### Features

- feat: changelog resource and data source @joshbeard (#100)

### Bug Fixes

- fix: doc attribute inconsistencies @joshbeard (#101)

### Maintenance

- build(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.4.2 to 1.5.0 @dependabot (#99)
- build(deps): bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 @dependabot (#98)
- build(deps): bump github.com/go-git/go-git/v5 from 5.10.1 to 5.11.0 @dependabot (#96)
- build(deps): bump crazy-max/ghaction-import-gpg from 6.0.0 to 6.1.0 @dependabot (#97)
- build(deps): bump golang.org/x/crypto from 0.15.0 to 0.17.0 @dependabot (#95)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.30.0 to 2.31.0 @dependabot (#94)
- build(deps): bump github.com/hashicorp/terraform-plugin-go from 0.19.1 to 0.20.0 @dependabot (#93)
- build(deps): bump github/codeql-action from 2 to 3 @dependabot (#92)
- build(deps): bump actions/setup-go from 4.1.0 to 5.0.0 @dependabot (#91)

## v0.2.1 - 2023-12-04

### Bug Fixes

- fix: api spec response error @joshbeard (#90)

## v0.2.0 - 2023-12-04

### Features

- feat: ability to associate doc with slug @joshbeard (#88)

### Bug Fixes

- fix: volatile 'user' attribute on docs @joshbeard (#87)

### Maintenance

- ci: test against Terraform 1.6 @joshbeard (#89)
- build(deps): bump github.com/hashicorp/terraform-plugin-go from 0.19.0 to 0.19.1 @dependabot (#86)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.29.0 to 2.30.0 @dependabot (#85)

## v0.1.17 - 2023-11-08

### Changes

- Update client to 0.1.3 for spec version bugfix @joshbeard (#84)

## v0.1.12 - 2023-04-18

### Features

- feat: Custom Pages data sources and resource @joshbeard (#27)

### Maintenance

- build(deps): bump mvdan.cc/gofumpt from 0.4.0 to 0.5.0 @dependabot (#26)
- build(deps): bump github.com/hashicorp/terraform-plugin-go from 0.14.3 to 0.15.0 @dependabot (#25)

## v0.1.11 - 2023-03-31

### Bug Fixes

- fix: Trim leading/trailing whitespace from docs @joshbeard (#23)

### Other Changes

- Docs/updates - examples, contributor workflow @joshbeard (#24)
- build(deps): bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.25.0 to 2.26.1 @dependabot (#22)
- build(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.1.1 to 1.2.0 @dependabot (#21)

## v0.1.10 - 2023-03-27

### Features

- feat: Sort api_specifications data source @joshbeard (#20)

## v0.1.9 - 2023-03-24

### Features

- feat: Add api_specifications data source @joshbeard (#19)

## v0.1.8 - 2023-03-20

### Changes

### Features

- feat: API spec data source filtering @joshbeard (#18)

## v0.1.7 - 2023-03-16

### Bug Fixes

- fix: api spec data source - lookup by title @joshbeard (#17)

## v0.1.6 - 2023-03-07

### Changes

### Bug Fixes

- fix: Don't send conflicting request params @joshbeard (#14)

## v0.1.5 - 2023-03-06

### Bug Fixes

- fix: image path validation @joshbeard (#11)

## v0.1.4 - 2023-03-02

### Changes

### Features

- feature: image upload @joshbeard (#10)

## v0.1.3 - 2023-02-28

### Changes

### Bug Fixes

- fix: re-create deleted resources @joshbeard (#9)

## v0.1.2 - 2023-02-23

### Changes

- fix: update registry provider address @joshbeard (#8)

### Maintenance

- ci: explicit file list for goreleaser @joshbeard (#7)

## v0.1.1 - 2023-02-22

### Changes

- Update package name and URL @joshbeard (#6)

## v0.1.0 - 2023-02-22

### Changes

- Initialize codebase @joshbeard (#1, #2, #5)
