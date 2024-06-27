package providers

func WithNodes(nodes interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Nodes = nodes.(uint)
	}
}

func WithNuma(numa interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Numa = numa.(uint)
	}
}

func WithMemory(memory interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Memory = memory.(string)
	}
}

func WithCPU(cpu interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.CPU = cpu.(uint)
	}
}

func WithSwap(swap interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Swap = swap.(bool)
	}
}

func WithUnlimitedSwap(us interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.UnlimitedSwap = us.(bool)
	}
}

func WithSwapiness(s interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Swapiness = s.(uint)
	}
}

func WithSwapSize(s interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Swapsize = s.(string)
	}
}

func WithSecondaryNics(secondaryNics interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SecondaryNics = secondaryNics.(uint)
	}
}

func WithQemuArgs(qemuArgs interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.QemuArgs = qemuArgs.(string)
	}
}

func WithKernelArgs(kernelArgs interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KernelArgs = kernelArgs.(string)
	}
}

func WithBackground(background interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Background = background.(bool)
	}
}

func WithRandomPorts(randomPorts interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RandomPorts = randomPorts.(bool)
	}
}

func WithSlim(slim interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Slim = slim.(bool)
	}
}

func WithVNCPort(vncPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.VNCPort = vncPort.(uint16)
	}
}

func WithHTTPPort(httpPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPPort = httpPort.(uint16)
	}
}

func WithHTTPSPort(httpsPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPSPort = httpsPort.(uint16)
	}
}

func WithRegistryPort(registryPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RegistryPort = registryPort.(uint16)
	}
}

func WithOCPort(ocpPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.OCPort = ocpPort.(uint16)
	}
}

func WithK8sPort(k8sPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.K8sPort = k8sPort.(uint16)
	}
}

func WithSSHPort(sshPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SSHPort = sshPort.(uint16)
	}
}

func WithPrometheusPort(prometheusPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.PrometheusPort = prometheusPort.(uint16)
	}
}

func WithGrafanaPort(grafanaPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GrafanaPort = grafanaPort.(uint16)
	}
}

func WithDNSPort(dnsPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DNSPort = dnsPort.(uint16)
	}
}

func WithNFSData(nfsData interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NFSData = nfsData.(string)
	}
}

func WithEnableCeph(enableCeph interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCeph = enableCeph.(bool)
	}
}

func WithEnableIstio(enableIstio interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableIstio = enableIstio.(bool)
	}
}

func WithEnableCNAO(enableCNAO interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCNAO = enableCNAO.(bool)
	}
}

func WithEnableNFSCSI(enableNFSCSI interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableNFSCSI = enableNFSCSI.(bool)
	}
}

func WithEnablePrometheus(enablePrometheus interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheus = enablePrometheus.(bool)
	}
}

func WithEnablePrometheusAlertManager(enablePrometheusAlertManager interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheusAlertManager = enablePrometheusAlertManager.(bool)
	}
}

func WithEnableGrafana(enableGrafana interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableGrafana = enableGrafana.(bool)
	}
}
func WithMultus(multus interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableMultus = multus.(bool)
	}
}
func WithAAQ(aaq interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.AAQ = aaq.(bool)
	}
}
func WithCDI(cdi interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.CDI = cdi.(bool)
	}
}

func WithKSM(ksm interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KSM = ksm.(bool)
	}
}

func WithKSMInterval(ki interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KSMInterval = ki.(uint)
	}
}

func WithKSMPages(kp interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KSMPages = kp.(uint)
	}
}

func WithDockerProxy(dockerProxy interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DockerProxy = dockerProxy.(string)
	}
}

func WithGPU(gpu interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GPU = gpu.(string)
	}
}

func WithCDIVersion(cdi interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.CDIVersion = cdi.(string)
	}
}

func WithAAQVersion(aaq interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.AAQVersion = aaq.(string)
	}
}

func WithNvmeDisks(nvmeDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NvmeDisks = nvmeDisks.([]string)
	}
}

func WithScsiDisks(scsiDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.ScsiDisks = scsiDisks.([]string)
	}
}

func WithRunEtcdOnMemory(runEtcdOnMemory interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RunEtcdOnMemory = runEtcdOnMemory.(bool)
	}
}

func WithEtcdCapacity(etcdCapacity interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EtcdCapacity = etcdCapacity.(string)
	}
}

func WithHugepages2M(hugepages2M interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Hugepages2M = hugepages2M.(uint)
	}
}

func WithEnableRealtimeScheduler(enableRealtimeScheduler interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableRealtimeScheduler = enableRealtimeScheduler.(bool)
	}
}

func WithEnableFIPS(enableFIPS interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableFIPS = enableFIPS.(bool)
	}
}

func WithEnablePSA(enablePSA interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePSA = enablePSA.(bool)
	}
}

func WithSingleStack(singleStack interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SingleStack = singleStack.(bool)
	}
}

func WithEnableAudit(enableAudit interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableAudit = enableAudit.(bool)
	}
}

func WithUSBDisks(usbDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.USBDisks = usbDisks.([]string)
	}
}
