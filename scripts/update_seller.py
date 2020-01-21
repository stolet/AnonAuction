import subprocess
import paramiko
import json
from datetime import datetime, timedelta

from constants import *

result = subprocess.run([(
        "az vmss list-instance-public-ips --resource-group %s --name %s --query [].{ip:ipAddress} -o tsv" % (
    ResourceGroup, VMSSname))], stdout=subprocess.PIPE, shell=True)

publicIPs = result.stdout.decode("utf-8").split("\n")[0:-1]

commands = []

def createSeller(ip):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(ip, username=VMusername, password=VMpassword)
    
    time = datetime.utcnow().replace(microsecond=0) +timedelta(minutes=2)

    config = json.dumps({'Item': "Fancy chocolate",
    'Auctioneers': [s + ":80" for s in publicIPs[1:]],
    'Interval': "30s",
    'T': 3,
    'StartTime': time.isoformat()+"Z"}, sort_keys=True,indent=4, separators=(',', ': '))

    ftp = ssh.open_sftp()
    file=ftp.file('P2-d3w9a-b3c0b-b3l0b-k0b9/seller/config.json', "w", -1)
    file.write(config)
    file.flush()
    ftp.close()


    ssh.close()

createSeller(publicIPs[0])