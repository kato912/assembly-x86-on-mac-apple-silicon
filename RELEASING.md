# Releasing nasmx86

One-time setup, then every release is just "push a tag".

## One-time setup

1. **Make this repository public** (Settings → General → Danger Zone →
   Change visibility). Homebrew and `install.sh` download the release files
   anonymously, which only works for public repositories.

2. **Create the tap repository** `kato912/homebrew-tap`:
   GitHub → New repository → name `homebrew-tap`, **Public**, tick
   "Add a README file" → Create. (The name must start with `homebrew-`; users
   then type `brew install kato912/tap/nasmx86`.)

3. **Create a token that may write to the tap** (no Homebrew account needed):
   GitHub → your avatar → Settings → Developer settings → Personal access
   tokens → **Fine-grained tokens** → Generate new token
   - Token name: `nasmx86 homebrew tap`
   - Expiration: e.g. 1 year (renew it when it expires)
   - Resource owner: `kato912`
   - Repository access: **Only select repositories** → `kato912/homebrew-tap`
   - Permissions → Repository permissions → **Contents: Read and write**
   - Generate, and copy the token (it is shown only once).

4. **Store the token as a secret of this repository**:
   this repo → Settings → Secrets and variables → Actions →
   New repository secret
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Secret: the token from step 3

## Every release

```bash
git switch main && git pull
go test ./...                       # optional local check
git tag -a v1.0.0 -m "nasmx86 v1.0.0"
git push origin v1.0.0
```

The **Release** workflow (Actions tab) then runs the tests, builds
darwin/arm64, darwin/amd64, linux/amd64, linux/arm64 and windows/amd64,
uploads them with `checksums.txt` to a GitHub Release, and commits
`Casks/nasmx86.rb` to `kato912/homebrew-tap`. (It is a cask, not a formula:
Homebrew requires an up-to-date Xcode to install a formula without bottles,
but not to install a cask.)

## Check the release

```bash
brew update && brew install kato912/tap/nasmx86 && nasmx86 --version   # macOS
curl -fsSL https://raw.githubusercontent.com/kato912/assembly-x86-on-mac-apple-silicon/main/install.sh | sh
```

If the Release workflow fails at the Homebrew step, check that the secret
name is exactly `HOMEBREW_TAP_GITHUB_TOKEN` and the token has Contents
read/write on `homebrew-tap`. To retry the same version: delete the release
and the tag on GitHub (`git push --delete origin v1.0.0`, `git tag -d v1.0.0`),
then tag and push again.
