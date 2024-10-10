#!/bin/bash

# Define log directory and file inside the workDir
LOG_DIR="logs"
LOG_FILE="create_video_$(date +%Y-%m-%d).log"
MAX_LOG_DAYS=5

# Function to print debug information in green
debug_print() {
    echo -e "\033[0;32m[DEBUG] $(date +"%Y-%m-%d %H:%M:%S") - $1\033[0m" | tee -a "$LOG_DIR/$LOG_FILE"
}

# Function to delete old log files
delete_old_logs() {
    find "$LOG_DIR" -name "create_video_*.log" -type f -mtime +$MAX_LOG_DAYS -delete
    debug_print "Deleted log files older than $MAX_LOG_DAYS days"
}

# Move to the ffmeg dir and print an error message if it fails
if ! cd ./ffmeg; then
    debug_print "Error: Failed to change directory to ./ffmeg"
    exit 1
fi

# Delete old logs
delete_old_logs

debug_print "---------------------------------------"
debug_print "Script started"

# Check if the workDir is provided as an argument
if [ -z "$1" ]; then
    echo "Usage: $0 <workDir>"
    exit 1
fi

# Set the workDir variable from the input argument
workDir=$1

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Get the duration of the audio file from the provided workDir
audio_file="${workDir}/audio.mp3"
if [ ! -f "$audio_file" ]; then
    debug_print "Audio file not found: $audio_file"
    exit 1
fi

debug_print "Getting audio duration..."
duration=$(./ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$audio_file")
debug_print "Audio duration: $duration seconds"

# Count the number of PNG images in the workDir
debug_print "Counting images..."
image_count=$(ls -1 ${workDir}/*.png 2>/dev/null | wc -l)
if [ "$image_count" -eq 0 ]; then
    debug_print "No PNG images found in $workDir"
    exit 1
fi
debug_print "Number of images: $image_count"

# Calculate the framerate and round to 6 decimal places
debug_print "Calculating framerate..."
framerate=$(echo "$image_count / $duration" | bc -l)
framerate=$(printf "%.6f" "$framerate")
debug_print "Calculated framerate (rounded): $framerate"

# Run the FFmpeg command with the calculated framerate
output_video="${workDir}/output.mp4"
debug_print "Running FFmpeg command..."
./ffmpeg -y -framerate $framerate -pattern_type glob -i "${workDir}/*.png" -i "$audio_file" -c:v libx264 -pix_fmt yuv420p -shortest "$output_video"

debug_print "FFmpeg command completed"
debug_print "Output video: $output_video"

debug_print "Script finished"
