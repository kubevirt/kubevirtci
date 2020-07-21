# pack8s kubevirtci provider status

kubevirtci `56f69bb5867db7517f70a0787b32570a861e124a`

## Supported providers:

| Provider          | Run           | Provisioning  | Notes              |
| ----------------- | ------------- | ------------- | ------------------ |
| k8s-1.14.6        | Yes           | Planned       |                    |
| k8s-1.15.1        | Yes           | Planned       |                    |
| k8s-1.16.2        | Yes           | Planned       |                    |
| k8s-multus-1.13.3 | Yes           | N/A           |                    |
| os-3.11.0         | Yes           | N/A           |                    |
| os-3.11.0-crio    | Yes           | N/A           |                    |
| os-3.11.0-multus  | Yes           | N/A           |                    |

Key:
- Yes: works like gocli, no regression known
- No: something's broken, see notes
- N/A: we don't plan to implement this
- In progress: the team started working on it, still WIP
- Not yet: planned for the near future, work has not begun yet
- Planned: planned for the far future, blocked by something else (likely the "not yet" queue)

## Unsupported providers:

* local
* k8s-1.11.0
* k8s-1.13.3
* kind
* kind-k8s-1.17
* kind-k8s-sriov-1.17.0
