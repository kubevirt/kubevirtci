# Provision Manager

## Purpose:
Detect which files are changed versus the latest `KUBEVIRTCI_TAG`,
and build only the providers which should be updated.
The other providers will be reused by retagging their latest version.

## Motivation:
1. Freeze a state unless a change happens.
2. Faster provision job.
3. Reduce collision domain, a problem on one provider won't block updating the others.

## Related Files:
1. `hack/pman/rules.yaml` - the rules that determine which folder/file affects which provider.
2. `hack/pman/force` - changing this file allows to enforce rebuild of all vm based providers.

## Flow:
1. Read rules.yaml and build a database of the rules.
2. For each changed file (comparing to latest `KUBEVIRTCI_TAG`), check which rule match:  
a. Check the full filename (relative to kubevirtci folder).  
b. Check the `dirname` of the file  
c. Check each of the `dirname`/* of the file and its parent directories (until `.`, not included).  
If a match is found, the assicated rule is accumulated.
Match must be found, unless the file is deleted,
because we have rules which are regex, so once some files, are deleted also their rule will be gone.  
The motivation is to have less to maintain when adding / removing new providers.

See `Config` struct at `cluster-provision/gocli/cmd/provision_manager.go` for more details about rule types.

## Parameters:
`--debug` - Run in debug mode.
Output has the following sections:
1. Run Parameters such as the tag comparing to, and the targets detected.
2. The expanded rules (based on rules file).
3. The files that were detected, their git status, and what is their effect based on the rules.
4. The cumulative targets (which is the output in case of non DEBUG mode as well).
Example: `docker run --rm -v $(pwd):/workdir:Z quay.io/kubevirtci/gocli provision-manager --debug`

## Notes:
1. If you need to enforce provision of all providers, run and commit this file:
`curl -sL https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest?ignoreCache=1 > hack/pman/force`
This is helpful in case you want a specific PR to rebuild all providers, without changing the publish job.
2. In order to bypass provision manager and rebuild all targets, run `BYPASS_PMAN=true ./publish.sh`.
