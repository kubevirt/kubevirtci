package providers

func WithNodes(nodes uint) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Nodes = nodes
	}
}

func WithNuma(numa uint) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Numa = numa
	}
}

func WithMemory(memory string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Memory = memory
	}
}

func WithCPU(cpu uint) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.CPU = cpu
	}
}

func WithSecondaryNics(secondaryNics uint) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SecondaryNics = secondaryNics
	}
}

func WithQemuArgs(qemuArgs string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.QemuArgs = qemuArgs
	}
}

func WithKernelArgs(kernelArgs string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KernelArgs = kernelArgs
	}
}

func WithBackground(background bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Background = background
	}
}

func WithReverse(reverse bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Reverse = reverse
	}
}

func WithRandomPorts(randomPorts bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RandomPorts = randomPorts
	}
}

func WithSlim(slim bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Slim = slim
	}
}

func WithVNCPort(vncPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.VNCPort = vncPort
	}
}

func WithHTTPPort(httpPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPPort = httpPort
	}
}

func WithHTTPSPort(httpsPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPSPort = httpsPort
	}
}

func WithRegistryPort(registryPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RegistryPort = registryPort
	}
}

func WithOCPort(ocpPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.OCPort = ocpPort
	}
}

func WithK8sPort(k8sPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.K8sPort = k8sPort
	}
}

func WithSSHPort(sshPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SSHPort = sshPort
	}
}

func WithPrometheusPort(prometheusPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.PrometheusPort = prometheusPort
	}
}

func WithGrafanaPort(grafanaPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GrafanaPort = grafanaPort
	}
}

func WithDNSPort(dnsPort uint16) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DNSPort = dnsPort
	}
}

func WithNFSData(nfsData string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NFSData = nfsData
	}
}

func WithEnableCeph(enableCeph bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCeph = enableCeph
	}
}

func WithEnableIstio(enableIstio bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableIstio = enableIstio
	}
}

func WithEnableCNAO(enableCNAO bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCNAO = enableCNAO
	}
}

func WithEnableNFSCSI(enableNFSCSI bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableNFSCSI = enableNFSCSI
	}
}

func WithEnablePrometheus(enablePrometheus bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheus = enablePrometheus
	}
}

func WithEnablePrometheusAlertManager(enablePrometheusAlertManager bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheusAlertManager = enablePrometheusAlertManager
	}
}

func WithEnableGrafana(enableGrafana bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableGrafana = enableGrafana
	}
}

func WithDockerProxy(dockerProxy string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DockerProxy = dockerProxy
	}
}

func WithGPU(gpu string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GPU = gpu
	}
}

func WithNvmeDisks(nvmeDisks []string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NvmeDisks = nvmeDisks
	}
}

func WithScsiDisks(scsiDisks []string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.ScsiDisks = scsiDisks
	}
}

func WithRunEtcdOnMemory(runEtcdOnMemory bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RunEtcdOnMemory = runEtcdOnMemory
	}
}

func WithEtcdCapacity(etcdCapacity string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EtcdCapacity = etcdCapacity
	}
}

func WithHugepages2M(hugepages2M uint) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Hugepages2M = hugepages2M
	}
}

func WithEnableRealtimeScheduler(enableRealtimeScheduler bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableRealtimeScheduler = enableRealtimeScheduler
	}
}

func WithEnableFIPS(enableFIPS bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableFIPS = enableFIPS
	}
}

func WithEnablePSA(enablePSA bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePSA = enablePSA
	}
}

func WithSingleStack(singleStack bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SingleStack = singleStack
	}
}

func WithEnableAudit(enableAudit bool) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableAudit = enableAudit
	}
}

func WithUSBDisks(usbDisks []string) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.USBDisks = usbDisks
	}
}
