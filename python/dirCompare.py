import os
import hashlib
from datetime import datetime, timedelta
import sys
import time
import itertools

# Define files and extensions to ignore
IGNORE_FILES = {'.DS_Store', 'Thumbs.db', 'desktop.ini', '.git', '.gitignore', '.gitattributes'}
IGNORE_EXTENSIONS = {'.tmp', '.bak', '.swp', '.swo', '.old', '.orig'}

def compute_file_checksum(file_path, hash_algorithm='sha256'):
    """
    Computes the checksum of a file.
    Args:
        file_path (str): Path to the file.
        hash_algorithm (str): The hash algorithm to use (default is SHA256).
    Returns:
        str: Hexadecimal checksum of the file.
    """
    hash_func = getattr(hashlib, hash_algorithm)()
    with open(file_path, 'rb') as f:
        while chunk := f.read(8192):  # Read in 8 KB chunks
            hash_func.update(chunk)
    return hash_func.hexdigest()

def compute_folder_checksum(folder_path, hash_algorithm='sha256'):
    """
    Computes the checksum of a folder and its contents, excluding ignored files.
    Args:
        folder_path (str): Path to the folder.
        hash_algorithm (str): The hash algorithm to use (default is SHA256).
    Returns:
        dict: A dictionary mapping file paths (relative to folder_path) to their checksums.
    """
    checksums = {}
    for root, _, files in os.walk(folder_path):
        for file in sorted(files):
            # Skip ignored files and extensions
            if file in IGNORE_FILES or any(file.endswith(ext) for ext in IGNORE_EXTENSIONS):
                continue
            file_path = os.path.join(root, file)
            relative_path = os.path.relpath(file_path, folder_path)
            checksums[relative_path] = compute_file_checksum(file_path, hash_algorithm)
    return checksums

def calculate_folder_size(folder_path):
    """
    Calculates the total size of a folder (recursively).
    Args:
        folder_path (str): Path to the folder.
    Returns:
        int: Total size in bytes.
    """
    total_size = 0
    for root, _, files in os.walk(folder_path):
        for file in files:
            file_path = os.path.join(root, file)
            if os.path.exists(file_path):  # Ensure the file exists
                total_size += os.path.getsize(file_path)
    return total_size

def format_size(bytes_size):
    """
    Formats the size from bytes to a human-readable format (e.g., KB, MB, GB).
    Args:
        bytes_size (int): Size in bytes.
    Returns:
        str: Human-readable size (e.g., '369MB').
    """
    for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
        if bytes_size < 1024.0:
            return f"{bytes_size:.0f}{unit}"
        bytes_size /= 1024.0

def format_duration(duration):
    """
    Formats a timedelta duration to a human-readable format.
    Args:
        duration (timedelta): The duration to format.
    Returns:
        str: Human-readable duration (e.g., '2 hours, 3 minutes, 45 seconds').
    """
    total_seconds = int(duration.total_seconds())
    hours, remainder = divmod(total_seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    if hours > 0:
        return f"{hours} hours, {minutes} minutes, {seconds} seconds"
    elif minutes > 0:
        return f"{minutes} minutes, {seconds} seconds"
    else:
        return f"{seconds} seconds"

def compare_folders(folder1_checksums, folder2_checksums):
    """
    Compares two folders' checksums.
    Args:
        folder1_checksums (dict): Checksums of the first folder.
        folder2_checksums (dict): Checksums of the second folder.
    Returns:
        dict: A comparison result including added, removed, and changed files.
    """
    folder1_files = set(folder1_checksums.keys())
    folder2_files = set(folder2_checksums.keys())

    added = folder2_files - folder1_files
    removed = folder1_files - folder2_files
    common = folder1_files & folder2_files

    changed = {
        file for file in common
        if folder1_checksums[file] != folder2_checksums[file]
    }

    return {
        "added": sorted(added),
        "removed": sorted(removed),
        "changed": sorted(changed),
    }

def read_config(file_path):
    """
    Reads the configuration file to extract directory paths.
    Args:
        file_path (str): Path to the configuration file.
    Returns:
        tuple: A tuple containing the original and destination directory paths.
    """
    original_dir = None
    destination_dir = None
    
    with open(file_path, 'r') as f:
        for line in f:
            if line.startswith("original="):
                original_dir = line.split("=", 1)[1].strip().strip('"')
            elif line.startswith("destination="):
                destination_dir = line.split("=", 1)[1].strip().strip('"')

    if not original_dir or not destination_dir:
        raise ValueError("Both original and destination directories must be specified in the config file.")

    return original_dir, destination_dir

def write_results_to_file(results, output_file, timestamp, duration, origin_size, destination_size):
    """
    Writes the comparison results to a file.
    Args:
        results (dict): The comparison results.
        output_file (str): Path to the output file.
        timestamp (str): Human-readable timestamp of when the script was run.
        duration (str): Time it took to complete the task.
        origin_size (str): Size of the origin folder in human-readable format.
        destination_size (str): Size of the destination folder in human-readable format.
    """
    # Ensure the output directory exists
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    
    with open(output_file, 'w') as f:
        f.write(f"Run at {timestamp}\n")
        f.write(f"Time to complete - {duration}\n\n")
        f.write("Size\n")
        f.write("--------\n")
        f.write(f"Origin Folder: {origin_size}\n")
        f.write(f"Destination Folder: {destination_size}\n\n")
        f.write("Comparison Results\n")
        f.write("--------------------\n")
        f.write(f"Added Files: {results['added']}\n")
        f.write(f"Removed Files: {results['removed']}\n")
        f.write(f"Changed Files: {results['changed']}\n")

def spinner():
    """Simple terminal spinner for progress indication."""
    for char in itertools.cycle(['|', '/', '-', '\\']):
        sys.stdout.write(f'\rProcessing {char}')
        sys.stdout.flush()
        time.sleep(0.1)

def main():
    """
    Main function to read the configuration file, compare directories, and output results to a file.
    """
    # Accept the configuration file as a command-line argument
    if len(sys.argv) != 2:
        print("Usage: python3 dirCompare.py <config-file>")
        return

    config_file = sys.argv[1]

    try:
        # Read the configuration file
        original_dir, destination_dir = read_config(config_file)
        
        # Validate directories
        if not os.path.exists(original_dir):
            raise FileNotFoundError(f"Original directory does not exist: {original_dir}")
        if not os.path.exists(destination_dir):
            raise FileNotFoundError(f"Destination directory does not exist: {destination_dir}")
        
        print(f"Comparing folders:\nOriginal: {original_dir}\nDestination: {destination_dir}")
        
        # Extract the base name of the original folder for the output file name
        original_folder_name = os.path.basename(os.path.normpath(original_dir))
        output_file = f"out/{original_folder_name}.out"
        
        # Get the current timestamp and start time
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        start_time = datetime.now()
        
        # Start the spinner in a separate thread
        import threading
        spinner_thread = threading.Thread(target=spinner, daemon=True)
        spinner_thread.start()
        
        # Calculate folder sizes
        origin_size = format_size(calculate_folder_size(original_dir))
        destination_size = format_size(calculate_folder_size(destination_dir))
        
        # Compute checksums for both directories
        original_checksums = compute_folder_checksum(original_dir)
        destination_checksums = compute_folder_checksum(destination_dir)
        
        # Compare the folders
        comparison = compare_folders(original_checksums, destination_checksums)
        
        # Stop the spinner
        spinner_thread.join(0)
        sys.stdout.write('\rProcessing complete!\n')
        sys.stdout.flush()
        
        # Calculate the total duration
        elapsed_time = datetime.now() - start_time
        duration = format_duration(elapsed_time)
        
        # Write results to the output file
        write_results_to_file(comparison, output_file, timestamp, duration, origin_size, destination_size)
        print(f"Comparison results written to: {output_file}")
        
    except FileNotFoundError as e:
        print(f"Error: {e}")
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    main()