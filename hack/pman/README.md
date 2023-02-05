# Provision Manager

## Purpose:
Detect which files are changed versus the latest `KUBEVIRTCI_TAG`,
and build only the providers which should be updated.
The other providers will be reused by retagging their latest version.

## Motivation:
1. Freeze a state unless a change happens.
2. Faster provision job.
3. Reduce collision domain, a problem on one provider won't block updating the others.

## Files:
1. `hack/pman/pman.sh` - the script itself.
2. `hack/pman/rules.txt` - the rules that determine which folder/file affects which provider.

## Flow:
1. Read rules.txt and build a database of the rules.
Line format: `directory - rule`
Rules type: (see more comments on rules.txt)  
`all` - all the vm based providers will be provisioned.  
`none` - none of the vm based providers will be provisioned.  
`value` - a specific name of vm based provider that will be provisioned.  
`regex` - the regex will be globbed, and for each directory there will be a rule
        where the directory affects the specific provider: `a/b/k8s-X.YZ - X.YZ`.  
`regex_none` - the regex will be globbed, and for each directory there will be a rule
        where the directory affects none of the providers: `cluster-up/cluster/kind-X.YZ - none`.  
`!value` - all beside `value` will be provisioned (exclude).  
2. For each changed file (comparing to `KUBEVIRTCI_TAG`), check which rule match:  
a. Check the full filename (relative to kubevirtci folder).  
b. Check the `dirname` of the file  
c. Check each of the `dirname`/* of the file and its parent directories (until `.` not included).  
If a match is found, the assicated rule is accumulated.
Match must be found, unless the file is deleted,
because we have rules which are regex, so once some files, are deleted also their rule will be gone.  
The motivaton is to less have to maintain it when adding / removing new providers.

## Parameters:
`DEBUG` - Run in debug mode.
Output has 3 sections:
1. The expanded rules.
2. The files that were detected and what is their effect.
3. The cumulative targets (which is the output in case of non DEBUG mode as well).
Example: `DEBUG=true ./hack/pman/pman.sh`

`OVERRIDE` - Override Provision Manager and enforce specific targets.
Values:
1. false (default) - Let pman decide.
2. all - print all targets (according `cluster-provision/k8s/*`)
3. "value1[ value2...]" - list of specific targets (i.e `OVERRIDE=1.24`, `OVERRIDE=1.24 1.25`)

## Notes:
1. If you need to enforce provision of all providers, run and commit this file:
`curl -sL https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest?ignoreCache=1 > hack/pman/force`
2. Changing `pman.sh` / `rules.txt` won't issue a rebuild (unless you change the rule of it).