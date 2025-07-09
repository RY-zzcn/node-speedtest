// 全局变量
const API_BASE_URL = '/api';
let currentUser = null;
let nodes = [];
let speedtestResults = [];
let settings = {};

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', function() {
    // 初始化Alpine.js数据
    initAlpineData();
    
    // 检查用户登录状态
    checkLoginStatus();
});

// 初始化Alpine.js数据
function initAlpineData() {
    Alpine.data('app', () => ({
        // 页面状态
        currentPage: 'dashboard',
        isLoading: false,
        showSidebar: true,
        darkMode: localStorage.getItem('darkMode') === 'true',
        
        // 用户相关
        isLoggedIn: false,
        user: null,
        loginForm: {
            username: '',
            password: '',
            error: ''
        },
        
        // 节点相关
        nodes: [],
        selectedNode: null,
        nodeForm: {
            id: '',
            name: '',
            ip: '',
            location: '',
            description: '',
            tags: []
        },
        showNodeModal: false,
        nodeModalMode: 'add', // 'add' 或 'edit'
        showInstallCommandModal: false,
        installCommand: '',
        
        // 测速相关
        speedtestResults: [],
        speedtestForm: {
            sourceNodeId: '',
            targetNodeId: '',
            type: 'full'
        },
        showSpeedtestModal: false,
        
        // 系统设置
        settings: {},
        settingsForm: {},
        
        // 统计信息
        stats: {
            onlineNodes: 0,
            offlineNodes: 0,
            totalNodes: 0,
            todayTests: 0,
            totalTests: 0,
            cpuUsage: 0,
            memoryUsage: 0,
            diskUsage: 0
        },
        
        // 初始化
        init() {
            // 监听登录状态变化
            this.$watch('isLoggedIn', (value) => {
                if (value) {
                    this.loadDashboard();
                } else {
                    this.currentPage = 'login';
                }
            });
            
            // 监听暗黑模式变化
            this.$watch('darkMode', (value) => {
                localStorage.setItem('darkMode', value);
                if (value) {
                    document.documentElement.classList.add('dark');
                } else {
                    document.documentElement.classList.remove('dark');
                }
            });
            
            // 初始化暗黑模式
            if (this.darkMode) {
                document.documentElement.classList.add('dark');
            }
        },
        
        // 切换页面
        changePage(page) {
            this.currentPage = page;
            
            // 根据页面加载数据
            switch (page) {
                case 'dashboard':
                    this.loadDashboard();
                    break;
                case 'nodes':
                    this.loadNodes();
                    break;
                case 'speedtest':
                    this.loadSpeedtestResults();
                    break;
                case 'settings':
                    this.loadSettings();
                    break;
            }
        },
        
        // 登录相关方法
        async login() {
            this.isLoading = true;
            this.loginForm.error = '';
            
            try {
                const response = await fetch(`${API_BASE_URL}/login`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: this.loginForm.username,
                        password: this.loginForm.password
                    })
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 登录成功
                    this.user = data.data.user;
                    this.isLoggedIn = true;
                    currentUser = data.data.user;
                    
                    // 保存令牌到本地存储
                    localStorage.setItem('token', data.data.token);
                    
                    // 切换到仪表板页面
                    this.changePage('dashboard');
                } else {
                    // 登录失败
                    this.loginForm.error = data.message || '用户名或密码错误';
                }
            } catch (error) {
                console.error('登录请求失败:', error);
                this.loginForm.error = '登录请求失败，请稍后再试';
            } finally {
                this.isLoading = false;
            }
        },
        
        async logout() {
            this.isLoading = true;
            
            try {
                await fetch(`${API_BASE_URL}/logout`, {
                    method: 'POST',
                    headers: getHeaders()
                });
                
                // 清除登录状态
                this.user = null;
                this.isLoggedIn = false;
                currentUser = null;
                localStorage.removeItem('token');
                
                // 切换到登录页面
                this.currentPage = 'login';
            } catch (error) {
                console.error('登出请求失败:', error);
            } finally {
                this.isLoading = false;
            }
        },
        
        // 仪表板相关方法
        async loadDashboard() {
            this.isLoading = true;
            
            try {
                // 加载统计信息
                const response = await fetch(`${API_BASE_URL}/stats`, {
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    this.stats = data.data;
                }
                
                // 加载节点列表
                await this.loadNodes();
                
                // 加载最近的测速结果
                await this.loadSpeedtestResults();
            } catch (error) {
                console.error('加载仪表板数据失败:', error);
            } finally {
                this.isLoading = false;
            }
        },
        
        // 节点相关方法
        async loadNodes() {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/nodes`, {
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    this.nodes = data.data.nodes;
                    nodes = data.data.nodes;
                }
            } catch (error) {
                console.error('加载节点列表失败:', error);
            } finally {
                this.isLoading = false;
            }
        },
        
        openAddNodeModal() {
            this.nodeForm = {
                id: '',
                name: '',
                ip: '',
                location: '',
                description: '',
                tags: []
            };
            this.nodeModalMode = 'add';
            this.showNodeModal = true;
        },
        
        openEditNodeModal(node) {
            this.nodeForm = {
                id: node.id,
                name: node.name,
                ip: node.ip,
                location: node.location || '',
                description: node.description || '',
                tags: node.tags || []
            };
            this.nodeModalMode = 'edit';
            this.showNodeModal = true;
        },
        
        async saveNode() {
            this.isLoading = true;
            
            try {
                const isEdit = this.nodeModalMode === 'edit';
                const url = isEdit ? `${API_BASE_URL}/nodes/${this.nodeForm.id}` : `${API_BASE_URL}/nodes`;
                const method = isEdit ? 'PUT' : 'POST';
                
                const response = await fetch(url, {
                    method: method,
                    headers: getHeaders(),
                    body: JSON.stringify(this.nodeForm)
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 保存成功，重新加载节点列表
                    this.showNodeModal = false;
                    await this.loadNodes();
                } else {
                    console.error('保存节点失败:', data.message);
                    alert(`保存节点失败: ${data.message}`);
                }
            } catch (error) {
                console.error('保存节点请求失败:', error);
                alert('保存节点请求失败，请稍后再试');
            } finally {
                this.isLoading = false;
            }
        },
        
        async deleteNode(nodeId) {
            if (!confirm('确定要删除此节点吗？')) {
                return;
            }
            
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/nodes/${nodeId}`, {
                    method: 'DELETE',
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 删除成功，重新加载节点列表
                    await this.loadNodes();
                } else {
                    console.error('删除节点失败:', data.message);
                    alert(`删除节点失败: ${data.message}`);
                }
            } catch (error) {
                console.error('删除节点请求失败:', error);
                alert('删除节点请求失败，请稍后再试');
            } finally {
                this.isLoading = false;
            }
        },
        
        // 生成节点安装命令
        async generateInstallCommand(nodeId) {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/nodes/${nodeId}/install-command`, {
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 显示安装命令
                    this.showInstallCommandModal = true;
                    this.installCommand = data.data.command;
                } else {
                    console.error('生成安装命令失败:', data.message);
                    alert(`生成安装命令失败: ${data.message}`);
                }
            } catch (error) {
                console.error('生成安装命令请求失败:', error);
                alert('生成安装命令请求失败，请稍后再试');
            } finally {
                this.isLoading = false;
            }
        },
        
        // 测速相关方法
        async loadSpeedtestResults() {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/speedtest/results`, {
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    this.speedtestResults = data.data.results;
                    speedtestResults = data.data.results;
                }
            } catch (error) {
                console.error('加载测速结果失败:', error);
            } finally {
                this.isLoading = false;
            }
        },
        
        openSpeedtestModal() {
            this.speedtestForm = {
                sourceNodeId: this.nodes.length > 0 ? this.nodes[0].id : '',
                targetNodeId: this.nodes.length > 1 ? this.nodes[1].id : '',
                type: 'full'
            };
            this.showSpeedtestModal = true;
        },
        
        async startSpeedtest() {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/speedtest`, {
                    method: 'POST',
                    headers: getHeaders(),
                    body: JSON.stringify({
                        sourceNodeId: this.speedtestForm.sourceNodeId,
                        targetNodeId: this.speedtestForm.targetNodeId,
                        type: this.speedtestForm.type
                    })
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 测速任务已创建，关闭模态框
                    this.showSpeedtestModal = false;
                    
                    // 提示用户
                    alert('测速任务已创建，请稍后查看结果');
                    
                    // 重新加载测速结果
                    setTimeout(() => this.loadSpeedtestResults(), 3000);
                } else {
                    console.error('创建测速任务失败:', data.message);
                    alert(`创建测速任务失败: ${data.message}`);
                }
            } catch (error) {
                console.error('创建测速任务请求失败:', error);
                alert('创建测速任务请求失败，请稍后再试');
            } finally {
                this.isLoading = false;
            }
        },
        
        // 设置相关方法
        async loadSettings() {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/settings`, {
                    headers: getHeaders()
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    this.settings = data.data;
                    this.settingsForm = { ...data.data };
                    settings = data.data;
                }
            } catch (error) {
                console.error('加载系统设置失败:', error);
            } finally {
                this.isLoading = false;
            }
        },
        
        async saveSettings() {
            this.isLoading = true;
            
            try {
                const response = await fetch(`${API_BASE_URL}/settings`, {
                    method: 'PUT',
                    headers: getHeaders(),
                    body: JSON.stringify(this.settingsForm)
                });
                
                const data = await response.json();
                
                if (data.code === 0) {
                    // 保存成功，更新设置
                    this.settings = { ...this.settingsForm };
                    settings = { ...this.settingsForm };
                    alert('设置已保存');
                } else {
                    console.error('保存设置失败:', data.message);
                    alert(`保存设置失败: ${data.message}`);
                }
            } catch (error) {
                console.error('保存设置请求失败:', error);
                alert('保存设置请求失败，请稍后再试');
            } finally {
                this.isLoading = false;
            }
        },
        
        // 工具方法
        formatDate(dateString) {
            if (!dateString) return '';
            const date = new Date(dateString);
            return date.toLocaleString();
        },
        
        formatSpeed(speed) {
            if (speed === undefined || speed === null) return '-';
            if (speed < 1) return (speed * 1000).toFixed(2) + ' Kbps';
            return speed.toFixed(2) + ' Mbps';
        },
        
        formatPing(ping) {
            if (ping === undefined || ping === null) return '-';
            return ping.toFixed(2) + ' ms';
        },
        
        getNodeStatusClass(status) {
            switch (status) {
                case 'online': return 'bg-green-500';
                case 'offline': return 'bg-red-500';
                case 'unknown': return 'bg-gray-500';
                default: return 'bg-gray-500';
            }
        },
        
        getSpeedtestStatusClass(status) {
            switch (status) {
                case 'completed': return 'text-green-500';
                case 'pending': return 'text-yellow-500';
                case 'running': return 'text-blue-500';
                case 'failed': return 'text-red-500';
                case 'timeout': return 'text-red-500';
                default: return 'text-gray-500';
            }
        },
        
        getNodeName(nodeId) {
            const node = this.nodes.find(n => n.id === nodeId);
            return node ? node.name : nodeId;
        }
    }));
}

// 检查用户登录状态
async function checkLoginStatus() {
    const token = localStorage.getItem('token');
    
    if (!token) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/user`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        const data = await response.json();
        
        if (data.code === 0) {
            // 用户已登录
            currentUser = data.data;
            
            // 更新Alpine.js数据
            const app = Alpine.store('app');
            if (app) {
                app.user = data.data;
                app.isLoggedIn = true;
            }
        } else {
            // 登录已过期
            localStorage.removeItem('token');
        }
    } catch (error) {
        console.error('检查登录状态失败:', error);
    }
}

// 获取请求头
function getHeaders() {
    const headers = {
        'Content-Type': 'application/json'
    };
    
    const token = localStorage.getItem('token');
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    return headers;
} 