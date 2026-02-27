import psutil
import os
import time
import logging
from datetime import datetime

#Configure logging
logging.basicConfig(
    filename='process_alerts.log',
    level=logging.WARNING,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

#Declare threshold variables
CPU_THRESHOLD: float = 80.0
MEM_THRESHOLD: float = 80.0

# Processes that shouldnt be killed
KILL_WHITELIST = {"kernel", "systemd", "init", "sshd", "python", "python3"}

def get_top_processes(n, sort_by):
    # print('do you want to sort by CPU or memory')
    # user_input=input("cpu or mem: ")
    
    processes=[]
    for process in psutil.process_iter(['pid', 'name', 'cpu_percent', 'memory_percent', 'memory_info', 'status']):
        try:
            processes.append(process.info)
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            pass
    
    key = 'cpu_percent' if sort_by=='cpu' else 'memory_percent'

    def get_value(process):
        value= process[key] # get the cpu or memory value from the tuple
        if value is None:
            value= 0
        return value
    sorted_processes = sorted(processes, key=get_value, reverse=True)

    top_n_processes = sorted_processes[:n]
    return top_n_processes

def kill_processes(process):
    """Attempt to gracefully terminate, then force-kill a runaway process."""
    pid = process['pid']
    name = process['name']

    if name.lower() in KILL_WHITELIST:
        msg = f"SKIPPED kill for whitelisted process: {name} (PID {pid})"
        logging.warning(msg)
        return msg

    try:
        proc = psutil.Process(pid)
        proc.terminate()          # SIGTERM — asks nicely first
        proc.wait(timeout=3)      # give it 3 seconds to exit
        msg = f"KILLED (SIGTERM) PID {pid} ({name})"
    except psutil.TimeoutExpired:
        proc.kill()               # SIGKILL — force kill if it didn't respond
        msg = f"KILLED (SIGKILL) PID {pid} ({name})"
    except (psutil.NoSuchProcess, psutil.AccessDenied) as e:
        msg = f"Could not kill PID {pid} ({name}): {e}"

    logging.warning(msg)
    return msg

def log_alert(process, reason) -> str:
    msg: str = (f"ALERT | PID: {process['pid']} | Name: {process['name']} | "
           f"CPU: {process['cpu_percent']:.1f}% | MEM: {process['memory_percent']:.1f}% | "
           f"Reason: {reason}")
    logging.warning(msg)
    return msg

def check_thresholds(processes) -> list:
    alerts: list = []
    for p in processes:
        if p['cpu_percent'] and p['cpu_percent'] > CPU_THRESHOLD:
            alerts.append(log_alert(p, f"CPU exceeded {CPU_THRESHOLD}%"))
        if p['memory_percent'] and p['memory_percent'] > MEM_THRESHOLD:
            alerts.append(log_alert(p, f"MEM exceeded {MEM_THRESHOLD}%"))
    return alerts

def display(processes, interval, alerts):
    os.system("clear") # run clear command in the terminal

    #system-wide stats
    print (f" CPU Total: {psutil.cpu_percent()}% "
           f" CPU Times: {psutil.cpu_times()} seconds "
           f" CPU Count: {psutil.cpu_count()}% "
           f"Memory: {psutil.virtual_memory().percent}%   "
           f"Swap: {psutil.swap_memory().percent}%")
    
    if alerts:
        print(f"\n⚠️  {len(alerts)} ALERT(S) — see process_alerts.log")
        for a in alerts[-3:]:  # show last 3 alerts inline
            print(f"   {a}")
    
    print(f"{'='*70}")
    print(f"{'PID':>7}  {'NAME':<25} {'CPU%':>6}  {'MEM%':>6}  {'RSS MB':>8}  {'STATUS':<10}")
    print(f"{'-'*70}")

    for process in processes:
        if process['memory_info'] is not None: # check if memory info exists for this process
            rss_bytes = process['memory_info'].rss # get the raw memory value (in bytes)
            rss_mb = rss_bytes / (1024 * 1024) # convert bytes to megabytes
        else:
            rss_mb = 0                          # if no memory info, default to 0

        flag = " ⚠️" if (process['cpu_percent'] or 0) > CPU_THRESHOLD or (process['memory_percent'] or 0) > MEM_THRESHOLD else "" # show flag when it exceeds the threshold

        print(f"{process['pid']:>7}  {process['name'][:25]:<25} {process['cpu_percent']:>6.1f}  "
            f"{process['memory_percent']:>6.1f}  {rss_mb:>8.1f}  {process['status']:<10}{flag}")
        
def monitor(n=10, interval=3, sort_by='cpu'):
    print("Starting monitor... (Ctrl+C to stop)")

    # First call to cpu_percent returns 0.0, warm it up
    for _ in psutil.process_iter(['cpu_percent']):
        pass
    time.sleep(interval)

    while True:
        processes = get_top_processes(n=n, sort_by=sort_by)
        alerts = check_thresholds(processes)
        display(processes, interval, alerts)
        time.sleep(interval)



if __name__ == "__main__":
    print('do you want to sort by CPU or memory')
    user_input=input("cpu or mem: ").strip().lower()

    monitor(n=10, interval=3, sort_by=user_input)  # Change sort_by to "memory" if needed