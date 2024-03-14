# command encoding service written in Go using a minimal number of packages

## Overview

The task is to implement an algorithm that:

1. Takes a string array as input.
2. Initializes an empty tree per command.
3. Builds a tree starting from the bottom, taking the two lowest frequencies as leaves for the subtree, and then assigning this tree to the array/list (so the root node of a subtree is treated as an element with frequency being a sum of leaves frequencies).
The process continues until one tree is built with values only in leaves.
4. Generates binary codes for each string from the input array by counting the levels of the tree and assigning 0 for the left and 1 for the right node traversal. For example, 2 levels until the leaf node, and both to the left node generate the code 00.

This algorithm involves a tree where values are only stored inside leaves, and the least frequent values are stored at the lowest level of the tree (more nodes from the root).
It resembles a Huffman tree (Huffman coding problem), but it operates on symbols being strings inside the string array at the input.

## Implementation Steps

1. Take a string array as input.
2. Initialize an empty tree per command.
3. Build a tree starting from the bottom:
   - Calculate frequencies and store them in something like a hash table.
   - Utilize a priority queue (heap) to take two elements with the lowest frequencies in every step.
     (Pushing to the heap is O(logn), which is more efficient than sorting in each iteration of the tree-building loop.)
   - Traverse to each leaf and generate codes.

4. Generate codes for each string by counting the levels of the tree and assigning 0 for the left and 1 for the right node traversal.

## Usage


```bash
# be sure docker is installed https://docs.docker.com/engine/install/

# Clone this repository
git clone https://github.com/jpataflafla/go_huffman_coding

# go to the project folder
cd go_huffman_coding

# Build and start the app
docker compose build
# If there's an issue, check permissions as described in potential errors (if not using sudo), or run "newgrp docker" on Linux.

# Start the service (you can do it without -d for detached mode)
docker compose up -d

# Done. The listening port exposed to the hosting device is 80.  

```
The listening port exposed to the hosting device is 80.  
To use/test API, perform the following actions:  

To send command log:
**POST:**
- **Endpoint:** `localhost:80/commands`
- **Body:**
  ```json
  {
    "commands": ["LEFT", "GRAB", "LEFT", "BACK", "LEFT", "BACK", "LEFT"]
  }
  
  ```
  
To get code generated for the command (generated based on the most recently added command log):
**GET:**
- **Endpoint:** `localhost:80/rcr/{commandName}`
- **example:**
  `localhost:80/rcr/GRAB`  
  response:
  ```json
  {
    "rcr": "00"
  }
  ```

Due to current limitations and the simplicity of the system, queries always refer to the most recent list of commands.  Although the database can store historical command logs, and future updates may allow you to specify which log to generate command code for, in the demo version the database only stores the last 100 command logs.  

To view the command logs stored inside db, use:  
**GET:**
- **Endpoint:** `localhost:80/commands`  

To view the command codes stored inside db (codes are generated only once when requested and then stored), use:  
**GET:**
- **Endpoint:** `localhost:80/allCommandCodes`


