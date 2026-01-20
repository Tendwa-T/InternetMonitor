# InternetMonitor 🌐

View your internet latency and speed from your MenuBar

## Overview

InternetMonitor is a macOS application that provides real-time monitoring of your internet connection directly from your menu bar. Built with Go, it offers a lightweight and efficient way to keep track of your network performance without interrupting your workflow.

## Features

- 📊 **Real-time Monitoring**: Track your internet latency and connection speed in real-time
- 🖥️ **MenuBar Integration**: Quick access to network statistics without opening a separate application
- ⚡ **Lightweight**:  Built with Go for optimal performance and minimal resource usage
- 🎯 **Simple Interface**: Clean and intuitive display of network metrics

## Project Structure

```
InternetMonitor/
├── agent/          # Backend monitoring agent
└── mac-ui/         # macOS MenuBar user interface
```

## Components

### Agent
The agent component handles the core monitoring functionality, including: 
- Network latency measurement
- Connection speed testing
- Data collection and processing

### Mac UI
The mac-ui component provides: 
- MenuBar integration
- Visual display of network metrics
- User interaction and configuration options

## Technologies Used

- **Go**: Core monitoring agent and backend logic
- **macOS APIs**: MenuBar integration and system-level access
- **Network Programming**: Custom latency and speed measurement implementations

## Showcase Highlights

This project demonstrates: 
- Cross-component architecture design (agent + UI)
- Real-time data processing and visualization
- Native macOS application development
- Network programming and performance monitoring
- Clean separation of concerns between backend and frontend

## Roadmap

- [ ] Support for additional network metrics
- [ ] Historical data tracking and visualization
- [ ] Customizable alert thresholds
- [ ] Export functionality for network statistics
- [ ] Cross-platform support (Windows, Linux)

## Author

**Tendwa-T** - [GitHub Profile](https://github.com/Tendwa-T)

## Contact

For inquiries about this project, please reach out via [GitHub](https://github.com/Tendwa-T).

---

**Note**: This is a showcase project demonstrating network monitoring capabilities and macOS application development.