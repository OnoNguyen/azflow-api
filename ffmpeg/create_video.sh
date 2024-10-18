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

# Move to the ffmpeg dir and print an error message if it fails
if ! cd ./ffmpeg; then
    debug_print "Error: Failed to change directory to ./ffmpeg"
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

# Create temporary directory for intermediate videos
TRUNK_DIR="$workDir/trunks"
mkdir -p "$TRUNK_DIR"

# Function to clean up temporary files
cleanup() {
    debug_print "Cleaning up temporary files..."
    rm -rf "$TRUNK_DIR"
}

# Function to process each pair of audio and image files
process_trunk() {
    local audio_file="$1"
    local image_file="$2"
    local base_name="$3"

    # Get duration of the current audio file
    debug_print "Getting duration for $audio_file..."
    duration=$(./ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$audio_file")
    duration=$(printf "%.6f" "$duration")
    debug_print "Duration: $duration seconds"

    # Calculate new duration with 1 second of silence added
    new_duration=$(echo "$duration + 1" | bc)
    debug_print "New video duration: $new_duration seconds"

    # Create intermediate video for this pair
    intermediate_video="${TRUNK_DIR}/${base_name}_video.mp4"

    debug_print "Creating intermediate video for $audio_file and $image_file with 1 second of silence..."
    ./ffmpeg -y \
        -loop 1 -t "$new_duration" -i "$image_file" \
        -i "$audio_file" \
        -filter_complex "[1:a]apad=pad_dur=1[apadded]" \
        -map 0:v -map "[apadded]" \
        -c:v libx264 -pix_fmt yuv420p \
        -c:a aac -shortest \
        "$intermediate_video"
}

# Find and process matching audio and image files
debug_print "Finding matching audio and image files..."
for audio_file in "${workDir}"/*.mp3; do
    base_name=$(basename "$audio_file" .mp3)
    image_file="${workDir}/${base_name}.png"

    if [ ! -f "$image_file" ]; then
        debug_print "No matching image found for $audio_file, skipping..."
        continue
    fi

    # Process the audio and image pair
    process_trunk "$audio_file" "$image_file" "$base_name"
done

# List all intermediate videos
intermediate_videos=$(ls "${TRUNK_DIR}"/*_video.mp4)
if [ -z "$intermediate_videos" ]; then
    debug_print "No intermediate videos created, exiting..."
    exit 1
fi

# Concatenate all intermediate videos into one final video
concat_list="${TRUNK_DIR}/concat_list.txt"
for video in $intermediate_videos; do
    echo "file '$video'" >> "$concat_list"
done

output_video="${workDir}/output.mp4"
debug_print "Concatenating intermediate videos into $output_video..."
./ffmpeg -y -f concat -safe 0 -i "$concat_list" -c copy "$output_video"

# Compress the final output video
#compressed_video="${workDir}/output_compressed.mp4"
#debug_print "Compressing the final output video..."
#./ffmpeg -y -i "$output_video" -vcodec libx265 -crf 28 -preset fast "$compressed_video"
#
#debug_print "FFmpeg concatenation and compression completed"
#debug_print "Compressed output video: $compressed_video"

debug_print "Script finished"
