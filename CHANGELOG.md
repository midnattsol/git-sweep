# Changelog

## [2.2.0](https://github.com/midnattsol/git-sweep/compare/v2.1.1...v2.2.0) (2026-02-13)


### Features

* add protect/unprotect commands and simplify Bitbucket auth ([1fd2d11](https://github.com/midnattsol/git-sweep/commit/1fd2d11de0a88ac62cccd38fd25a1435cb4ef8ce))

## [2.1.1](https://github.com/midnattsol/git-sweep/compare/v2.1.0...v2.1.1) (2026-02-13)


### Bug Fixes

* use MergedAt instead of GetMerged for PR detection ([545d878](https://github.com/midnattsol/git-sweep/commit/545d87839f2b4f6f3fc3a3fa9450d5b0b5aed11b))

## [2.1.0](https://github.com/midnattsol/git-sweep/compare/v2.0.1...v2.1.0) (2026-02-13)


### Features

* add stash subcommand for cleaning old stashes ([67ae951](https://github.com/midnattsol/git-sweep/commit/67ae951c1ef34201a3b2871bff394d8cbca1ac51))

## [2.0.1](https://github.com/midnattsol/git-sweep/compare/v2.0.0...v2.0.1) (2026-02-13)


### Bug Fixes

* combine release-please and goreleaser in single workflow ([32f98ec](https://github.com/midnattsol/git-sweep/commit/32f98ec5a5fc5a5723fb7d995ccb84136b319d98))

## [2.0.0](https://github.com/midnattsol/git-sweep/compare/v1.3.0...v2.0.0) (2026-02-13)


### ⚠ BREAKING CHANGES

* Remove --execute flag

### Features

* new CLI UX with confirmation prompt and --candidates flag ([8e7468e](https://github.com/midnattsol/git-sweep/commit/8e7468eca0a92c367f8392e0f077fcf96eb766a6))


### Bug Fixes

* revert to default release-please behavior ([bef7692](https://github.com/midnattsol/git-sweep/commit/bef7692fbcba1cc0f51f031b9b3f3ae69f7a5410))
* trigger goreleaser on release created instead of tag push ([d4ef34a](https://github.com/midnattsol/git-sweep/commit/d4ef34a0fc4cf50b1ab1f86f6aa553557ee7d66d))

## [1.3.0](https://github.com/midnattsol/git-sweep/compare/v1.2.0...v1.3.0) (2026-02-13)


### Features

* add multi-provider support (GitHub, GitLab, Bitbucket) ([d38a653](https://github.com/midnattsol/git-sweep/commit/d38a6538fc7dcf74c715fc352dafce81a06998f2))


### Bug Fixes

* add actions write permission to release-please ([f635c11](https://github.com/midnattsol/git-sweep/commit/f635c11b896aa2709dce73069b58aea806cd24c3))
* configure release-please to skip github release creation ([483af2e](https://github.com/midnattsol/git-sweep/commit/483af2e537b8778cf12ecf0fa82578ea94439465))
* separate release-please and goreleaser workflows ([51ce432](https://github.com/midnattsol/git-sweep/commit/51ce432f9fd0a06b2761302d3d69587424c873aa))

## [1.2.0](https://github.com/midnattsol/git-sweep/compare/v1.1.1...v1.2.0) (2026-02-13)


### Features

* add smart nuke mode with categories, self-update, and UI improvements ([828672e](https://github.com/midnattsol/git-sweep/commit/828672e7a78e8ea49c7b329d357c4b75cefac795))


### Bug Fixes

* correct branch name in install.sh usage comment ([574830a](https://github.com/midnattsol/git-sweep/commit/574830a457f15c42deb74c6c771b59d888e15918))

## [1.1.1](https://github.com/midnattsol/git-sweep/compare/v1.1.0...v1.1.1) (2026-02-13)


### Bug Fixes

* remove goreleaser warnings and ignore docs in release ([f4de7b0](https://github.com/midnattsol/git-sweep/commit/f4de7b0d2df2d0bcf541bc87c573059a2e6c4e80))

## [1.1.0](https://github.com/midnattsol/git-sweep/compare/v1.0.0...v1.1.0) (2026-02-13)


### Features

* initial implementation of git-sweep CLI ([b9ac89f](https://github.com/midnattsol/git-sweep/commit/b9ac89f0db4dfba2ed5a4d579030c3ee9cd9e360))


### Bug Fixes

* release-please on develop branch ([1bb0928](https://github.com/midnattsol/git-sweep/commit/1bb0928ed58e869156c52cb4536ae07779aa0f7d))
* unitfy release workflow with goreleaser ([90cca1a](https://github.com/midnattsol/git-sweep/commit/90cca1af172101c70474b788426b2e82dd3ec518))

## 1.0.0 (2026-02-13)


### Features

* initial implementation of git-sweep CLI ([b9ac89f](https://github.com/midnattsol/git-sweep/commit/b9ac89f0db4dfba2ed5a4d579030c3ee9cd9e360))


### Bug Fixes

* release-please on develop branch ([1bb0928](https://github.com/midnattsol/git-sweep/commit/1bb0928ed58e869156c52cb4536ae07779aa0f7d))
