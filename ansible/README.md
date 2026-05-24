# Ansible OCI Host Configuration

Manual-first Ansible setup for post-provision configuration of the OCI instance managed by `terraform/oci`.

## What it configures

- OCI dynamic inventory via the Oracle OCI Ansible collection
- host-level iptables rules
- nginx
- certbot in webroot mode with a systemd renewal timer
- concierge bot build and systemd deployment

The current default certificate email is `luiz@justanother.engineer`, matching the public `abox` example layout. The nginx and certbot roles are intentionally modeled after the working `abox` nginx and `nginx_certbot` behavior.

Out of scope:

- Docker
- containerctl
- app builds or app deployment
- partitioning

## Layout

```text
ansible/
  ansible.cfg
  inventory/oci.oci.yml
  inventory/group_vars/all.yml
  playbooks/site.yml
  roles/
  requirements.yml
```

## Prerequisites

- Ansible installed locally
- OCI Python SDK installed locally
- OCI credentials configured via the normal OCI config/environment flow
- optional `TF_VAR_ssh_ingress_cidr` if you want the host firewall SSH allowlist to match Terraform; when unset, the Ansible host firewall defaults SSH ingress to `0.0.0.0/0`
- DNS for `oci.justanother.engineer` pointing at the OCI instance before certbot runs
- SSH access to the instance with the matching private key already available locally
- a trusted SSH host key for `oci.justanother.engineer`; Ansible connects to the OCI public IP but reuses the FQDN as the host-key alias

## Install collections

```sh
cd ansible
bash bootstrap.sh
```

That script:

- detects the Python interpreter used by `ansible-playbook`
- bootstraps `pip` into that interpreter if needed
- installs `requirements.txt` into the same interpreter
- installs the Ansible collection from `requirements.yml`

If you want to run the steps manually, use the interpreter behind `ansible-playbook`, not a generic `python3`:

```sh
cd ansible
ANSIBLE_PYTHON="$(head -n 1 "$(command -v ansible-playbook)" | sed 's/^#!//')"
"$ANSIBLE_PYTHON" -m ensurepip --upgrade
"$ANSIBLE_PYTHON" -m pip install -r requirements.txt
ansible-galaxy collection install -r requirements.yml
```

## Inspect inventory

```sh
cd ansible
ansible-inventory -i inventory/oci.oci.yml --list
```

## Run the playbook

```sh
cd ansible
ansible-playbook -i inventory/oci.oci.yml playbooks/site.yml
```

## Deploying the concierge bot

The `concierge` role builds `../src/cmd/concierge` on the Ansible controller for `linux/amd64`, copies the binary to `/opt/concierge/concierge`, renders `/etc/concierge/concierge.env` from Ansible variables, and manages `concierge.service` with systemd.

Provide bot runtime configuration through the `concierge_env` Ansible variable. Example:

```yaml
concierge_env:
  SLACK_APP_TOKEN: xapp-...
  SLACK_BOT_TOKEN: xoxb-...
  SLACK_REQUESTS_CHANNEL_ID: C12345678
  SLACK_USER_IDS: U111,U222
  SLACK_MANAGER_IDS: U333
  SLACK_ADMIN_IDS: U444
  GITHUB_APP_ID: "12345"
  GITHUB_APP_INSTALLATION_ID: "67890"
  GITHUB_APP_PRIVATE_KEY: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
  GITHUB_OWNER: jae-labs
  GITHUB_REPO: conCIerge
```

When passing `GITHUB_APP_PRIVATE_KEY` through the generated systemd env file, use `\n` escapes between PEM lines. The Go config loader normalizes those escapes back into real newlines at runtime.

## Check mode

```sh
cd ansible
ansible-playbook -i inventory/oci.oci.yml playbooks/site.yml --check
```
