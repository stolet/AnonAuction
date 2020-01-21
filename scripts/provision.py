import subprocess
import urllib
import paramiko
import json

from constants import *

result = subprocess.run([(
        "az vmss list-instance-public-ips --resource-group %s --name %s --query [].{ip:ipAddress} -o tsv" % (
    ResourceGroup, VMSSname))], stdout=subprocess.PIPE, shell=True)
publicIPs = result.stdout.decode("utf-8").split("\n")[0:-1]


def clone(ssh):
    ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command(
        (
                "git clone https://" + ubcUsername + ":" + urllib.parse.quote_plus(
            ubcPassword) + "@github.ugrad.cs.ubc.ca/CPSC416-2018W-T1/P2-d3w9a-b3c0b-b3l0b-k0b9.git"))
    ssh_stdout.channel.recv_exit_status()
    lines = ssh_stdout.readlines()
    for line in lines:
        print(line)

    lines = ssh_stderr.readlines()
    for line in lines:
        print(line)


def getHostname(ssh):
    ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command("hostname -i")
    ssh_stdout.channel.recv_exit_status()
    lines = ssh_stdout.readlines()
    return lines[0].split("\n")[0]


def createSeller(ip):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(ip, username=VMusername, password=VMpassword)
    clone(ssh)

    ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command("./P2-d3w9a-b3c0b-b3l0b-k0b9/setup.sh")
    ssh_stdout.channel.recv_exit_status()

    ssh.close()


createSeller(publicIPs[0])

for ip in publicIPs[1:]:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(ip, username=VMusername, password=VMpassword)

    clone(ssh)
    localIP = getHostname(ssh)

    jsonfile = json.dumps(
        {'SellerIpPort': publicIPs[0] + ":80", 'LocalIpPort': localIP + ":80", 'ExternalIpPort': ip + ":80"},
        sort_keys=True, indent=4, separators=(',', ': '))

    ftp = ssh.open_sftp()
    file = ftp.file('P2-d3w9a-b3c0b-b3l0b-k0b9/auctioneer/config.json', "w", -1)
    file.write(jsonfile)
    file.flush()
    ftp.close()

    ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command("./P2-d3w9a-b3c0b-b3l0b-k0b9/setup.sh")
    ssh_stdout.channel.recv_exit_status()

    ssh.close()

print("seller " + publicIPs[0])
print("auctioneers " + str(publicIPs[1:]))
