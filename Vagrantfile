# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|

  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = "api.proficionym.dev"
  
  config.vm.provider "virtualbox" do |v|
    v.memory = 1024
    v.cpus = 1
    end
  
  #config.vm.network "forwarded_port", guest: 3000, host: 80
  config.vm.network :private_network, :auto_network => true
  
  config.vm.synced_folder "./", "/home/vagrant/workspace/src/proficionym"
  
  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y build-essential git curl nginx whois
    curl -sSf https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz -o golang.tar.gz
    tar -C /usr/local -xzf golang.tar.gz
    chown -R vagrant /home/vagrant
  SHELL

  # add paths
  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    cat >> ~/.bashrc <<EOF
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/workspace
EOF
  SHELL
end
