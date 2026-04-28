# AgileOS BPM Platform - Azure Infrastructure
# Terraform configuration for Azure deployment

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# Variables
variable "resource_group_name" {
  description = "Name of the resource group"
  default     = "agileos-rg"
}

variable "location" {
  description = "Azure region"
  default     = "southeastasia"
}

variable "environment" {
  description = "Environment name"
  default     = "production"
}

# Resource Group
resource "azurerm_resource_group" "agileos" {
  name     = var.resource_group_name
  location = var.location

  tags = {
    Environment = var.environment
    Project     = "AgileOS-BPM"
  }
}

# Virtual Network
resource "azurerm_virtual_network" "agileos" {
  name                = "agileos-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.agileos.location
  resource_group_name = azurerm_resource_group.agileos.name

  tags = {
    Environment = var.environment
  }
}

# Subnet
resource "azurerm_subnet" "agileos" {
  name                 = "agileos-subnet"
  resource_group_name  = azurerm_resource_group.agileos.name
  virtual_network_name = azurerm_virtual_network.agileos.name
  address_prefixes     = ["10.0.1.0/24"]
}

# Network Security Group
resource "azurerm_network_security_group" "agileos" {
  name                = "agileos-nsg"
  location            = azurerm_resource_group.agileos.location
  resource_group_name = azurerm_resource_group.agileos.name

  security_rule {
    name                       = "HTTP"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "HTTPS"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "SSH"
    priority                   = 120
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = {
    Environment = var.environment
  }
}

# Public IP
resource "azurerm_public_ip" "agileos" {
  name                = "agileos-public-ip"
  location            = azurerm_resource_group.agileos.location
  resource_group_name = azurerm_resource_group.agileos.name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    Environment = var.environment
  }
}

# Network Interface
resource "azurerm_network_interface" "agileos" {
  name                = "agileos-nic"
  location            = azurerm_resource_group.agileos.location
  resource_group_name = azurerm_resource_group.agileos.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.agileos.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.agileos.id
  }

  tags = {
    Environment = var.environment
  }
}

# Associate NSG with NIC
resource "azurerm_network_interface_security_group_association" "agileos" {
  network_interface_id      = azurerm_network_interface.agileos.id
  network_security_group_id = azurerm_network_security_group.agileos.id
}

# Virtual Machine
resource "azurerm_linux_virtual_machine" "agileos" {
  name                = "agileos-vm"
  resource_group_name = azurerm_resource_group.agileos.name
  location            = azurerm_resource_group.agileos.location
  size                = "Standard_B2s"
  admin_username      = "azureuser"

  network_interface_ids = [
    azurerm_network_interface.agileos.id,
  ]

  admin_ssh_key {
    username   = "azureuser"
    public_key = file("~/.ssh/id_rsa.pub")
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
    disk_size_gb         = 64
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  custom_data = base64encode(<<-EOF
    #!/bin/bash
    
    # Update system
    apt-get update
    apt-get upgrade -y
    
    # Install Docker
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    usermod -aG docker azureuser
    
    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    # Install Azure CLI
    curl -sL https://aka.ms/InstallAzureCLIDeb | bash
    
    # Create app directory
    mkdir -p /home/azureuser/agileos
    chown azureuser:azureuser /home/azureuser/agileos
    
    # Enable Docker service
    systemctl enable docker
    systemctl start docker
  EOF
  )

  tags = {
    Environment = var.environment
  }
}

# Azure Container Registry
resource "azurerm_container_registry" "agileos" {
  name                = "agileosacr"
  resource_group_name = azurerm_resource_group.agileos.name
  location            = azurerm_resource_group.agileos.location
  sku                 = "Basic"
  admin_enabled       = true

  tags = {
    Environment = var.environment
  }
}

# Outputs
output "public_ip_address" {
  value       = azurerm_public_ip.agileos.ip_address
  description = "The public IP address of the VM"
}

output "acr_login_server" {
  value       = azurerm_container_registry.agileos.login_server
  description = "The login server URL for the container registry"
}

output "acr_admin_username" {
  value       = azurerm_container_registry.agileos.admin_username
  description = "The admin username for the container registry"
  sensitive   = true
}

output "acr_admin_password" {
  value       = azurerm_container_registry.agileos.admin_password
  description = "The admin password for the container registry"
  sensitive   = true
}
