import subprocess
import urllib
import paramiko
import json

from constants import *
result = subprocess.run([(
        "az vmss list-instance-public-ips --resource-group %s --name %s --query [].{ip:ipAddress} -o tsv" % (
    ResourceGroup, VMSSname))], stdout=subprocess.PIPE, shell=True)
publicIPs = result.stdout.decode("utf-8").split("\n")[0:-1]

def pull(ssh):
    ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command("cd P2-d3w9a-b3c0b-b3l0b-k0b9; git stash; git pull; git stash pop")
    ssh_stdout.channel.recv_exit_status()
    lines = ssh_stdout.readlines()
    for line in lines:
        print(line)

    lines = ssh_stderr.readlines()
    for line in lines:
        print(line)


for ip in publicIPs:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(ip, username=VMusername, password=VMpassword)

    pull(ssh)
    ssh.close()

print("seller " + publicIPs[0])
print("auctioneers " + str(publicIPs[1:]))

