VAGRANTFILE_API_VERSION = "2"

Vagrant.require_version ">= 1.6.0"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

    config.vm.box = "ubuntu/trusty64"
    config.vm.provider :virtualbox do |v|
        v.memory = 1024
        v.cpus = 1
    end

    # define project root
    project_gopath = "/home/project"
    project_root = "#{project_gopath}/src/zdd"

    config.vm.network "private_network", type: "dhcp"
    config.vm.network "forwarded_port", guest: 80, host: 8080

    config.vm.synced_folder ".", project_root, type: "nfs"
    config.vm.provision "shell",
        path: "ansible/bootstrap.sh",
        args: [project_gopath, project_root],
        keep_color: true
end
