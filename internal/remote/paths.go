package remote

import "inx/internal/config"

// defaultManagedKnownHosts is the Inx-managed known_hosts path. It is a
// thin indirection over config so tests can leave HostKeyPolicy.ManagedPath
// empty and still get an isolated file under INX_HOME.
func defaultManagedKnownHosts() string {
	return config.RemoteKnownHostsPath()
}
