import psutil
import os
import time


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


def display(processes, interval):
    os.system("clear") # run clear command in the terminal

    #system-wide stats
    print (f" CPU Total: {psutil.cpu_percent()}% "
           f" CPU Times: {psutil.cpu_times()} seconds "
           f" CPU Count: {psutil.cpu_count()}% "
           f"Memory: {psutil.virtual_memory().percent}%   "
           f"Swap: {psutil.swap_memory().percent}%")
    
    print(f"{'='*70}")
    print(f"{'PID':>7}  {'NAME':<25} {'CPU%':>6}  {'MEM%':>6}  {'RSS MB':>8}  {'STATUS':<10}")
    print(f"{'-'*70}")

    for process in processes:
        if process['memory_info'] is not None: # check if memory info exists for this process
            rss_bytes = process['memory_info'].rss # get the raw memory value (in bytes)
            rss_mb = rss_bytes / (1024 * 1024) # convert bytes to megabytes
        else:
            rss_mb = 0                          # if no memory info, default to 0

        print(f"{process['pid']:>7}  {process['name'][:25]:<25} {process['cpu_percent']:>6.1f}  "
            f"{process['memory_percent']:>6.1f}  {rss_mb:>8.1f}  {process['status']:<10}")
        
def monitor(n=10, interval=3, sort_by='cpu'):
    print("Starting monitor... (Ctrl+C to stop)")

    # First call to cpu_percent returns 0.0, warm it up
    for process in psutil.process_iter(['cpu_percent']):
        pass
    time.sleep(interval)



    while True:
        processes = get_top_processes(n=n, sort_by=sort_by)
        display(processes, interval)
        time.sleep(interval)

if __name__ == "__main__":

    print('do you want to sort by CPU or memory')
    user_input=input("cpu or mem: ")

    monitor(n=10, interval=3, sort_by=user_input)  # Change sort_by to "memory" if needed