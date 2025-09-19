<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <h3>HeroBox</h3>
      </div>
      
      <el-menu
        :default-active="$route.path"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>仪表板</span>
        </el-menu-item>
        
        <el-sub-menu index="/sing-box">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>Sing-Box管理</span>
          </template>
          <el-menu-item index="/sing-box/inbound">
            <el-icon><Download /></el-icon>
            <span>入站设置</span>
          </el-menu-item>
          <el-menu-item index="/sing-box/proxy">
            <el-icon><Share /></el-icon>
            <span>代理/出站</span>
          </el-menu-item>
          <el-menu-item index="/sing-box/rules">
            <el-icon><List /></el-icon>
            <span>规则设置</span>
          </el-menu-item>
          <el-menu-item index="/sing-box">
            <el-icon><Setting /></el-icon>
            <span>服务管理</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-menu-item index="/mosdns">
          <el-icon><Monitor /></el-icon>
          <span>MosDNS管理</span>
        </el-menu-item>
        
        <el-menu-item index="/logs">
          <el-icon><DocumentCopy /></el-icon>
          <span>日志查看</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <!-- 主内容区 -->
    <el-container>
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="$route.meta.title">{{ $route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <span class="system-info">
            <el-icon><Monitor /></el-icon>
            内网管理系统
          </span>
        </div>
      </el-header>
      
      <!-- 主内容 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { 
  Odometer, Connection, Monitor, DocumentCopy, Setting, Download, Share, List
} from '@element-plus/icons-vue'

// 移除认证相关逻辑，简化内网使用
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #434a59;
}

.header {
  background-color: white;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header-left {
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
}

.system-info {
  display: flex;
  align-items: center;
  color: #606266;
  font-size: 14px;
  gap: 6px;
}


.main-content {
  background-color: #f5f7fa;
  padding: 20px;
}

.el-menu {
  border-right: none;
}

.el-menu-item {
  height: 50px;
  line-height: 50px;
}

.el-menu-item:hover {
  background-color: #434a59 !important;
}

.el-menu-item.is-active {
  background-color: rgba(64, 158, 255, 0.3) !important;
  border-right: 3px solid #409eff;
}

.el-menu-item.is-active span,
.el-menu-item.is-active .el-icon {
  color: #ffffff !important;
  font-weight: 600;
}
</style>
