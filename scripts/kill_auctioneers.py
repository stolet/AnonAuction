import subprocess
import sys
import threading

import paramiko

from constants import *

result = subprocess.run([(
        "az vmss list-instance-public-ips --resource-group %s --name %s --query [].{ip:ipAddress} -o tsv" % (
    ResourceGroup, VMSSname))], stdout=subprocess.PIPE, shell=True)
publicIPs = result.stdout.decode("utf-8").split("\n")[0:-1]


class SSHThread(threading.Thread):
    def __init__(self, ip):
        super(SSHThread, self).__init__()
        self.ip = ip

    def run(self):
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(self.ip, username=VMusername, password=VMpassword)

        ssh_stdin, ssh_stdout, ssh_stderr = ssh.exec_command(
            "sudo pkill auctioneer_main")

        for line in iter(ssh_stdout.readline, ""):
            print(self.ip + ": " + line, end="")

        for line in iter(ssh_stderr.readline, ""):
            print(self.ip + ": " + line, end="")

        ssh.close()


threads = []
for i in range(int(sys.argv[1])):
    ssh_thread = SSHThread(publicIPs[i+1])
    threads.append(ssh_thread)
    ssh_thread.start()

for thread in threads:
    thread.join()
