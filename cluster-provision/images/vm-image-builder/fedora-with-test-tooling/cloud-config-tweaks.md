# Fedora Test-Tooling Cloud-Config Tweaks

- **Access and users**
  - Create users and set non-expiring passwords for `fedora` and `cloud-user`.
  - This provides stable login credentials for users and cloud automation.

- **Kernel/module readiness**
  - Auto-load SR-IOV-related Mellanox/Intel NIC modules at boot.
  - This allows binding and using VFIO devices on the guest.

- **Cloud-init and guest-agent ordering**
  - Add a wait service so the guest agent starts after cloud-init.
  - This way, waiting for guest agents guarantees that cloud-init has completed.
  - Enable both services.

- **Console stability for tests**
  - Disable bracketed-paste output noise.
  - Disable systemd shell integration escapes.
  - This allows cleaner shell login without noisy characters.

- **Faster datasource detection**
  - Limit the datasource list to `NoCloud`, `ConfigDrive`, and `None`.
  - This can reduce boot time, especially when cloud-init would otherwise probe unavailable datasources.

- **Network and resolver behavior**
  - Restore classic interface naming (`eth0`) - currently, tests depend on these names.
  - Set NSS hosts order to `files dns myhostname` - this supports FQDN in subdomain tests.

- **Packages and netcat behavior**
  - Install VM/e2e tooling (guest agent, stress, perf/network/debug utilities).
  - Keep `nc` as OpenBSD netcat for cloud-init; also install `ncat` (`nmap-ncat`) since KubeVirt e2e uses both.

- **Image/test compatibility adjustments**
  - Clear static + transient hostname.
  - Remove `users_groups` so cloud-init does not remove the configured users.
  - Set SELinux to permissive mode.
  - Remove `pam_nologin` from sshd PAM - this allows login without waiting for non-required modules to settle.

- **Architecture-specific boot safety**
  - On `s390x`, regenerate initramfs and boot artifacts - this fixes boot panics.
