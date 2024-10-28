#!/bin/bash

# Define log directory and file inside the work_dir
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

# Check if the work_dir is provided as an argument
if [ -z "$1" ]; then
    echo "Usage: $0 <work_dir>"
    exit 1
fi

# Set the work_dir variable from the input argument
work_dir=$1

quote="\"\""
# assign 2 arg to $quote if argument is provided
if [ -n "$2" ]; then
    quote="$2"
fi

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Create temporary directory for intermediate videos
TRUNK_DIR="$work_dir/trunks"
rm -rf "$TRUNK_DIR"
mkdir -p "$TRUNK_DIR"

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

    # Calculate new duration with 0.5 second of silence added
    new_duration=$(echo "$duration + 0.5" | bc)
    debug_print "New video duration: $new_duration seconds"

    # Format duration from seconds to HH:MM:SS.ms format
#    formatted_duration=$(printf "0:%02d:%02d.%02d" \
#        $((${duration%.*} / 60)) \
#        $((${duration%.*} % 60)) \
#        $((${duration#*.} / 10000)))

    # Create the ASS subtitle file by reading template and replacing placeholders
#    sed -e "s/\[Start\]/0:00:00.00/g" \
#        -e "s/\[End\]/${formatted_duration}/g" \
#        -e "s/\[Text\]/${quote}/g" \
#        template.ass > "${TRUNK_DIR}/${base_name}.ass"

    # Create intermediate video for this pair
    intermediate_video="${TRUNK_DIR}/${base_name}_video.mp4"
    debug_print "Creating intermediate video for $audio_file and $image_file..."

    # Check if subtitle file exists and set the filter_complex accordingly
    if [ -f "${work_dir}/${base_name}.ass" ]; then
        filter_complex="[1:a]apad=pad_dur=1[apadded]; \
                        [0:v]subtitles=${work_dir}/${base_name}.ass [video_with_subtitles]"
        map_video="[video_with_subtitles]"
    else
        filter_complex="[1:a]apad=pad_dur=1[apadded]"
        map_video="0:v"
    fi

    # Run ffmpeg command with the chosen filter_complex
    ./ffmpeg -y \
        -loop 1 -t "$new_duration" -i "$image_file" \
        -i "$audio_file" \
        -filter_complex "$filter_complex" \
        -map "$map_video" -map "[apadded]" \
        -c:v libx264 -pix_fmt yuv420p \
        -c:a aac -shortest \
        "$intermediate_video"

}

# Find and process matching audio and image files
debug_print "Finding matching audio and image files..."
last_used_image_file=""
for audio_file in "${work_dir}"/*.mp3; do
    base_name=$(basename "$audio_file" .mp3)
    image_file="${work_dir}/${base_name}.png"

    if [ ! -f "$image_file" ]; then
      if [ -f "$last_used_image_file" ]; then
        debug_print "No matching image found for $audio_file, reusing $last_used_image_file..."
        image_file="$last_used_image_file"
      else
        debug_print "No matching image found for $audio_file, exiting..."
        continue
      fi
    fi

    # Process the audio and image pair
    process_trunk "$audio_file" "$image_file" "$base_name" "$quote"

    # Update the last used image file
    last_used_image_file="$image_file"
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

output_video="${work_dir}/output.mp4"
debug_print "Concatenating intermediate videos into $output_video..."
./ffmpeg -y -f concat -safe 0 -i "$concat_list" -c copy "$output_video"

# Compress the final output video
#compressed_video="${work_dir}/output_compressed.mp4"
#debug_print "Compressing the final output video..."
#./ffmpeg -y -i "$output_video" -vcodec libx265 -crf 28 -preset fast "$compressed_video"
#
#debug_print "FFmpeg concatenation and compression completed"
#debug_print "Compressed output video: $compressed_video"

debug_print "Script finished"
