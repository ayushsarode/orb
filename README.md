# Orb

A custom implementation of the core Git version control system (`orb`) written in Go, and a web interface clone inspired by GitHub (`orbhub`) built with Next.js.

![image](https://github.com/user-attachments/assets/564ca51b-6b61-499b-959c-1aa35e7a7c47)


## Overview

This project is primarily an ambitious **learning exercise** aimed at deeply understanding:

1.  **Git Internals:** How version control systems like Git work under the hood (objects, DAGs, index, refs, protocols).


## Project Components

1.  **`orb` (Go CLI):**
    *   **Goal:** To replicate core Git functionalities from the command line.
    *   **Implementation:** Written purely in Go.
    *   **Focus:** Understanding and implementing the Git object model (blobs, trees, commits), content-addressable storage, the index (staging area), refs (branches, HEAD), and eventually networking protocols (HTTP Smart Protocol).


## Features (Planned / In Progress)

### `orb` (Go CLI) - MVP Goals

*   [x] `orb init`: Initialize a new `.orb` repository.
*   [x] `orb hash-object`: Create blob objects from files.
*   [ ] `orb cat-file`: Inspect Git objects (blob, tree, commit).
*   [x] `orb add`: Stage file changes (update the index).
*   [x] `orb write-tree`: Create a tree object from the index.
*   [x] `orb commit-tree`: Create a commit object.
*   [x] `orb commit`: Create a commit (combining `write-tree`, `commit-tree`, and updating refs).
*   [x] `orb log`: View commit history.
*   [x] `orb status`: Show working directory and staging area status.

### `orb` (Go CLI) - Future Goals

*   [x] Branching (`orb branch`, `orb checkout`)
*   [ ] Merging
*   [ ] Diffing (`orb diff`)
*   [ ] Networking (`orb clone`, `orb fetch`, `orb pull`, `orb push`) via HTTP Smart Protocol
*   [ ] Networking via SSH Protocol
*   [ ] Tagging (`orb tag`)
*   [ ] Garbage collection / Packing


## Tech Stack

*   **`orb`:**
    *   Language: **Go**
    *   Core Libraries: Go Standard Library (`os`, `crypto/sha1` or `sha256`, `compress/zlib`, `net/http`, etc.)
    *   CLI Framework: `cobra`

### Git Concepts I’ve Covered

- [x] Understand the Git object model: **blob**, **tree**, **commit**, **tag**
- [x] Learn how Git stores objects in `.git/objects` using **SHA-1 hashes**
- [x] Understand how **refs** (branches, tags) and **HEAD** work
- [x] Explore the **index (staging area)** and how it tracks changes
- [ ] Grasp the structure of the **commit graph (DAG)** and how Git traverses history



