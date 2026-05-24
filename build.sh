#!/bin/bash

# Geektime Downloader 多平台自动化编译脚本
# 支持 Windows、macOS、Linux 多个平台和架构

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
APP_NAME="geektime-downloader"
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
BUILD_DIR="dist"

# 支持的平台和架构
PLATFORMS=(
    "windows-amd64:windows:amd64:.exe"
    "windows-386:windows:386:.exe"
    "windows-arm64:windows:arm64:.exe"
    "linux-amd64:linux:amd64:"
    "linux-386:linux:386:"
    "linux-arm64:linux:arm64:"
    "linux-arm:linux:arm:"
    "darwin-amd64:darwin:amd64:"
    "darwin-arm64:darwin:arm64:"
)

# 函数：打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数：显示帮助信息
show_help() {
    echo "Geektime Downloader 多平台编译脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -v, --version           显示版本信息"
    echo "  -p, --platform PLATFORM 编译指定平台 (例如: windows-amd64)"
    echo "  -a, --all               编译所有支持的平台 (默认)"
    echo "  -c, --clean             清理构建目录"
    echo "  -l, --list              列出所有支持的平台"
    echo "  --no-compress           不压缩输出文件"
    echo "  --ldflags STRING        传递额外的 ldflags"
    echo ""
    echo "示例:"
    echo "  $0                      # 编译所有平台"
    echo "  $0 -p windows-amd64     # 只编译 Windows AMD64"
    echo "  $0 -p linux-amd64       # 只编译 Linux AMD64"
    echo "  $0 -c                   # 清理构建目录"
    echo ""
    echo "支持的平台:"
    for platform_info in "${PLATFORMS[@]}"; do
        IFS=':' read -r platform os arch ext <<< "$platform_info"
        echo "  $platform"
    done
}

# 函数：显示版本信息
show_version() {
    echo "Geektime Downloader Build Script"
    echo "Version: $VERSION"
    echo "Commit: $COMMIT_HASH"
    echo "Build Time: $BUILD_TIME"
}

# 函数：列出支持的平台
list_platforms() {
    print_info "支持的编译平台:"
    for platform_info in "${PLATFORMS[@]}"; do
        IFS=':' read -r platform os arch ext <<< "$platform_info"
        printf "  %-20s %s/%s\n" "$platform" "$os" "$arch"
    done
}

# 函数：清理构建目录
clean_build() {
    print_info "清理构建目录..."
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"
        print_success "构建目录已清理"
    else
        print_warning "构建目录不存在"
    fi
}

# 函数：检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装或未在 PATH 中"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_info "Go 版本: $GO_VERSION"
    
    # 检查 git (可选)
    if ! command -v git &> /dev/null; then
        print_warning "Git 未安装，将使用默认版本信息"
    fi
    
    # 检查 upx (可选)
    if command -v upx &> /dev/null && [ "$COMPRESS" = "true" ]; then
        UPX_VERSION=$(upx --version | head -n1)
        print_info "UPX 版本: $UPX_VERSION"
    elif [ "$COMPRESS" = "true" ]; then
        print_warning "UPX 未安装，将跳过压缩步骤"
        COMPRESS="false"
    fi
}

# 函数：准备构建环境
prepare_build() {
    print_info "准备构建环境..."
    
    # 创建构建目录
    mkdir -p "$BUILD_DIR"
    
    # 检查 go.mod 文件
    if [ ! -f "go.mod" ]; then
        print_error "未找到 go.mod 文件，请确保在项目根目录执行此脚本"
        exit 1
    fi
    
    # 下载依赖
    print_info "下载依赖..."
    go mod tidy
    go mod download
}

# 函数：构建指定平台
build_platform() {
    local target_platform=$1
    local found=false
    local os arch ext
    
    # 查找匹配的平台
    for platform_info in "${PLATFORMS[@]}"; do
        IFS=':' read -r platform platform_os platform_arch platform_ext <<< "$platform_info"
        if [ "$platform" = "$target_platform" ]; then
            os=$platform_os
            arch=$platform_arch
            ext=$platform_ext
            found=true
            break
        fi
    done
    
    if [ "$found" = false ]; then
        print_error "不支持的平台: $target_platform"
        return 1
    fi
    
    local output_name="${APP_NAME}-${VERSION}-${os}-${arch}${ext}"
    local output_path="${BUILD_DIR}/${output_name}"
    
    print_info "编译 ${target_platform} (${os}/${arch})..."
    
    # 设置环境变量
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=0
    
    # 构建 ldflags
    local ldflags="-s -w"
    ldflags="$ldflags -X 'main.Version=$VERSION'"
    ldflags="$ldflags -X 'main.CommitHash=$COMMIT_HASH'"
    ldflags="$ldflags -X 'main.BuildTime=$BUILD_TIME'"
    
    if [ -n "$EXTRA_LDFLAGS" ]; then
        ldflags="$ldflags $EXTRA_LDFLAGS"
    fi
    
    # 执行编译
    if go build -ldflags "$ldflags" -o "$output_path" .; then
        local file_size=$(du -h "$output_path" | cut -f1)
        print_success "编译完成: $output_name (大小: $file_size)"
        
        # 压缩文件（如果启用）
        if [ "$COMPRESS" = "true" ] && command -v upx &> /dev/null; then
            print_info "压缩 $output_name..."
            if upx --best --lzma "$output_path" &> /dev/null; then
                local compressed_size=$(du -h "$output_path" | cut -f1)
                print_success "压缩完成: $output_name (压缩后: $compressed_size)"
            else
                print_warning "压缩失败: $output_name"
            fi
        fi
        
        return 0
    else
        print_error "编译失败: $target_platform"
        return 1
    fi
}

# 函数：生成校验和
generate_checksums() {
    print_info "生成校验和文件..."
    local checksum_file="${BUILD_DIR}/checksums.txt"
    
    cd "$BUILD_DIR"
    
    # 生成 SHA256 校验和
    if command -v sha256sum &> /dev/null; then
        sha256sum * > checksums.txt
    elif command -v shasum &> /dev/null; then
        shasum -a 256 * > checksums.txt
    else
        print_warning "未找到 sha256sum 或 shasum 命令，跳过校验和生成"
        cd ..
        return
    fi
    
    cd ..
    print_success "校验和文件已生成: $checksum_file"
}

# 函数：显示构建结果
show_results() {
    print_info "构建结果:"
    
    if [ -d "$BUILD_DIR" ]; then
        local total_files=$(find "$BUILD_DIR" -name "${APP_NAME}*" -type f | wc -l)
        local total_size=$(du -sh "$BUILD_DIR" 2>/dev/null | cut -f1 || echo "未知")
        
        print_success "总共构建了 $total_files 个文件，总大小: $total_size"
        
        echo ""
        printf "%-30s %-10s %-15s\n" "文件名" "大小" "平台"
        printf "%-30s %-10s %-15s\n" "----" "----" "----"
        
        for file in "$BUILD_DIR"/${APP_NAME}*; do
            if [ -f "$file" ]; then
                local filename=$(basename "$file")
                local filesize=$(du -h "$file" | cut -f1)
                local platform=$(echo "$filename" | sed "s/${APP_NAME}-${VERSION}-//g" | sed 's/\.[^.]*$//')
                printf "%-30s %-10s %-15s\n" "$filename" "$filesize" "$platform"
            fi
        done
        
        echo ""
        print_info "所有文件保存在: $BUILD_DIR/"
        
        # 显示校验和文件
        if [ -f "${BUILD_DIR}/checksums.txt" ]; then
            print_info "校验和文件: ${BUILD_DIR}/checksums.txt"
        fi
    else
        print_warning "构建目录不存在"
    fi
}

# 主函数
main() {
    local build_all=true
    local target_platform=""
    local clean_only=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                show_version
                exit 0
                ;;
            -p|--platform)
                target_platform="$2"
                build_all=false
                shift 2
                ;;
            -a|--all)
                build_all=true
                shift
                ;;
            -c|--clean)
                clean_only=true
                shift
                ;;
            -l|--list)
                list_platforms
                exit 0
                ;;
            --no-compress)
                COMPRESS="false"
                shift
                ;;
            --ldflags)
                EXTRA_LDFLAGS="$2"
                shift 2
                ;;
            *)
                print_error "未知参数: $1"
                echo "使用 $0 --help 查看帮助信息"
                exit 1
                ;;
        esac
    done
    
    # 设置默认值
    COMPRESS=${COMPRESS:-"true"}
    
    # 显示脚本信息
    echo "========================================"
    echo "  Geektime Downloader 构建脚本"
    echo "========================================"
    echo "版本: $VERSION"
    echo "提交: $COMMIT_HASH"
    echo "时间: $BUILD_TIME"
    echo "========================================"
    
    # 如果只是清理，执行清理后退出
    if [ "$clean_only" = true ]; then
        clean_build
        exit 0
    fi
    
    # 检查依赖
    check_dependencies
    
    # 准备构建环境
    prepare_build
    
    # 开始构建
    local success_count=0
    local total_count=0
    
    if [ "$build_all" = true ]; then
        print_info "开始构建所有支持的平台..."
        for platform_info in "${PLATFORMS[@]}"; do
            IFS=':' read -r platform os arch ext <<< "$platform_info"
            ((total_count++))
            if build_platform "$platform"; then
                ((success_count++))
            fi
            echo ""
        done
    else
        if [ -n "$target_platform" ]; then
            print_info "开始构建平台: $target_platform"
            ((total_count++))
            if build_platform "$target_platform"; then
                ((success_count++))
            fi
        else
            print_error "未指定目标平台"
            exit 1
        fi
    fi
    
    # 生成校验和
    if [ $success_count -gt 0 ]; then
        generate_checksums
    fi
    
    # 显示构建结果
    echo ""
    echo "========================================"
    print_info "构建完成! 成功: $success_count/$total_count"
    echo "========================================"
    show_results
    
    # 设置退出码
    if [ $success_count -eq $total_count ]; then
        exit 0
    else
        exit 1
    fi
}

# 执行主函数
main "$@"