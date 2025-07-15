#!/bin/bash

# 设置应用自动启动的脚本
# 支持 Ubuntu 和 Arch 系统
# 会在当前用户下设置 app 程序自动启动

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # 无颜色

# 应用程序路径
APP_PATH=""
# 工作目录
WORKING_DIR=""
# 应用程序名称
APP_NAME="backup-go"
# 系统类型
SYSTEM_TYPE=""

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    command -v $1 >/dev/null 2>&1 || {
        log_error "需要 $1 命令，但未找到。请先安装。"
        exit 1
    }
}

# 确定系统类型
determine_system() {
    if [ -f "/etc/os-release" ]; then
        . /etc/os-release
        case "$ID" in
        ubuntu)
            SYSTEM_TYPE="ubuntu"
            ;;
        arch)
            SYSTEM_TYPE="arch"
            ;;
        *)
            log_error "不支持的系统类型: $ID"
            exit 1
            ;;
        esac
    else
        log_error "无法确定系统类型"
        exit 1
    fi
    log_info "检测到系统: $SYSTEM_TYPE"
}

# 获取应用路径
get_app_path() {
    while [ -z "$APP_PATH" ]; do
        read -p "请输入 $APP_NAME 程序的完整路径: " APP_PATH
        if [ ! -f "$APP_PATH" ]; then
            log_error "文件不存在: $APP_PATH"
            APP_PATH=""
        elif [ ! -x "$APP_PATH" ]; then
            log_warn "文件不可执行，尝试添加执行权限..."
            chmod +x "$APP_PATH"
            if [ $? -ne 0 ]; then
                log_error "无法添加执行权限，请检查文件权限"
                APP_PATH=""
            else
                log_info "已成功添加执行权限"
            fi
        fi
    done
    log_info "应用路径: $APP_PATH"
}

# 获取工作目录
get_working_dir() {
    while [ -z "$WORKING_DIR" ]; do
        read -p "请输入 $APP_NAME 程序的工作目录 (直接回车使用程序所在目录): " WORKING_DIR_TEMP
        if [ -z "$WORKING_DIR_TEMP" ]; then
            WORKING_DIR=$(dirname "$APP_PATH")
        elif [ ! -d "$WORKING_DIR_TEMP" ]; then
            log_error "目录不存在: $WORKING_DIR_TEMP"
            WORKING_DIR=""
        else
            WORKING_DIR="$WORKING_DIR_TEMP"
        fi
    done
    log_info "工作目录: $WORKING_DIR"
}

# 创建并配置systemd服务（公共逻辑）
setup_common_service() {
    log_info "开始设置 $APP_NAME 服务..."

    # 创建 systemd 服务文件
    SERVICE_FILE="$HOME/.config/systemd/user/$APP_NAME.service"

    log_info "创建服务文件: $SERVICE_FILE"
    cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=$APP_NAME 服务
After=network.target

[Service]
Type=simple
ExecStart=$APP_PATH
WorkingDirectory=$WORKING_DIR
Restart=always
RestartSec=5s

[Install]
WantedBy=default.target
EOF

    # 重新加载 systemd
    log_info "重新加载 systemd 用户实例"
    systemctl --user daemon-reload

    # 启用服务
    log_info "启用 $APP_NAME 服务"
    systemctl --user enable $APP_NAME.service

    # 启动服务
    log_info "启动 $APP_NAME 服务"
    systemctl --user start $APP_NAME.service

    # 检查服务状态
    log_info "检查 $APP_NAME 服务状态"
    systemctl --user status $APP_NAME.service
}

# Ubuntu系统特有的设置
setup_ubuntu_specific() {
    log_info "Ubuntu 系统自动启动设置完成！"
}

# Arch系统特有的设置
setup_arch_specific() {
    # 确保用户登录时启动服务
    log_info "设置用户登录时自动启动服务"
    loginctl enable-linger $(whoami)

    log_info "Arch 系统自动启动设置完成！"
}

# 在 Ubuntu 系统上设置自动启动
setup_ubuntu() {
    setup_common_service
    setup_ubuntu_specific
    log_info "服务控制命令:"
    log_info "  启动: systemctl --user start $APP_NAME.service"
    log_info "  停止: systemctl --user stop $APP_NAME.service"
    log_info "  重启: systemctl --user restart $APP_NAME.service"
    log_info "  状态: systemctl --user status $APP_NAME.service"
}

# 在 Arch 系统上设置自动启动
setup_arch() {
    setup_common_service
    setup_arch_specific
    log_info "服务控制命令:"
    log_info "  启动: systemctl --user start $APP_NAME.service"
    log_info "  停止: systemctl --user stop $APP_NAME.service"
    log_info "  重启: systemctl --user restart $APP_NAME.service"
    log_info "  状态: systemctl --user status $APP_NAME.service"
}

# 主函数
main() {
    log_info "===== $APP_NAME 自动启动设置工具 ====="

    # 确定系统类型
    determine_system

    # 检查必要的命令
    check_command systemctl

    # 获取应用路径
    get_app_path

    # 获取工作目录
    get_working_dir

    # 根据系统类型执行相应的设置
    case "$SYSTEM_TYPE" in
    ubuntu)
        setup_ubuntu
        ;;
    arch)
        setup_arch
        ;;
    esac

    log_info "===== 设置完成 ====="
}

# 执行主函数
main
