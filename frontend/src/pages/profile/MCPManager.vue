<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Sparkles, Plus, CheckCircle, Clock, AlertCircle, XCircle, Copy, Shield } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import * as clipboard from 'clipboard-polyfill'

const auth = useAuthStore()

// 状态管理
const myKeys = ref<any[]>([])
const adminKeys = ref<any[]>([])
const activeTab = ref<'user' | 'admin'>('user')
const loading = ref(true)
const actionLoading = ref<string>('') // 记录哪个操作在 loading (Key ID)

// 申请表单
const form = ref({
	applicant: '',
	purpose: '',
})
const showApplyForm = ref(false)
const submitting = ref(false)

// 获取远程 SSE 地址
const getMcpUrl = (key: string) => {
	const host = window.location.origin
	return `${host}/api/v1/mcp/sse?api_key=${key}`
}

// 获取用户自己申请的 Keys
async function fetchMyKeys() {
	loading.value = true
	try {
		const res: any = await request.get('/user/mcp/keys/my')
		if (res.code === 0) {
			myKeys.value = res.data || []
		}
	} catch (error) {
		console.error('获取 MCP Key 失败:', error)
	} finally {
		loading.value = false
	}
}

// 获取管理员待审批列表 (只在管理员角色下调用)
async function fetchAdminKeys() {
	if (auth.user?.role !== 'admin') return
	loading.value = true
	try {
		const res: any = await request.get('/admin/mcp/keys')
		if (res.code === 0) {
			adminKeys.value = res.data || []
		}
	} catch (error) {
		console.error('管理员拉取 MCP Key 列表失败:', error)
	} finally {
		loading.value = false
	}
}

// 提交申请
async function handleApply() {
	if (!form.value.applicant || !form.value.purpose) {
		toast.error('请填写完整的申请信息')
		return
	}

	try {
		submitting.value = true
		const res: any = await request.post('/user/mcp/keys/apply', form.value)
		if (res.code === 0) {
			toast.success('申请提交成功，请等待管理员审批！')
			showApplyForm.value = false
			form.value.applicant = ''
			form.value.purpose = ''
			await fetchMyKeys()
		} else {
			toast.error(res.msg || '申请提交失败')
		}
	} catch (error: any) {
		toast.error(error.msg || '网络或未知错误')
	} finally {
		submitting.value = false
	}
}

// 管理员操作：批准
async function handleApprove(id: string) {
	try {
		actionLoading.value = id
		const res: any = await request.post(`/admin/mcp/keys/${id}/approve`)
		if (res.code === 0) {
			toast.success('已成功批准申请并生成密钥！')
			await fetchAdminKeys()
		} else {
			toast.error(res.msg || '操作失败')
		}
	} catch (error: any) {
		toast.error(error.msg || '审批失败')
	} finally {
		actionLoading.value = ''
	}
}

// 管理员操作：驳回
async function handleReject(id: string) {
	if (!confirm('确定要拒绝此申请吗？')) return
	try {
		actionLoading.value = id
		const res: any = await request.post(`/admin/mcp/keys/${id}/reject`)
		if (res.code === 0) {
			toast.success('已驳回该申请')
			await fetchAdminKeys()
		} else {
			toast.error(res.msg || '操作失败')
		}
	} catch (error: any) {
		toast.error(error.msg || '操作失败')
	} finally {
		actionLoading.value = ''
	}
}

// 管理员操作：吊销
async function handleRevoke(id: string) {
	if (!confirm('确定要吊销此 API Key 吗？吊销后外部 Client 将立即无法连接！')) return
	try {
		actionLoading.value = id
		const res: any = await request.post(`/admin/mcp/keys/${id}/revoke`)
		if (res.code === 0) {
			toast.success('已成功吊销该密钥')
			if (activeTab.value === 'admin') {
				await fetchAdminKeys()
			} else {
				await fetchMyKeys()
			}
		} else {
			toast.error(res.msg || '操作失败')
		}
	} catch (error: any) {
		toast.error(error.msg || '操作失败')
	} finally {
		actionLoading.value = ''
	}
}

// 复制到剪贴板
function copyText(text: string) {
	clipboard.writeText(text).then(() => {
		toast.success('复制成功！')
	}).catch(() => {
		toast.error('复制失败，请长按手动选择复制')
	})
}

// 初始化
onMounted(() => {
	fetchMyKeys()
	if (auth.user?.role === 'admin') {
		fetchAdminKeys()
	}
})

// 监听 tab 切换
function switchTab(tab: 'user' | 'admin') {
	activeTab.value = tab
	if (tab === 'user') {
		fetchMyKeys()
	} else {
		fetchAdminKeys()
	}
}
</script>

<template>
	<div class="mcp-page">
		<TopNavBar title="AI 智能助理集成 (MCP)" />

		<div class="page-content">
			<!-- 管理员 Tab 切换 -->
			<div class="tab-header" v-if="auth.user?.role === 'admin'">
				<button 
					class="tab-btn" 
					:class="{ active: activeTab === 'user' }" 
					@click="switchTab('user')"
				>
					我的集成
				</button>
				<button 
					class="tab-btn" 
					:class="{ active: activeTab === 'admin' }" 
					@click="switchTab('admin')"
				>
					<Shield :size="14" style="margin-right: 4px; vertical-align: middle;" /> 审批管理
				</button>
			</div>

			<!-- Tab 1: 普通用户 - 我的集成 -->
			<div v-if="activeTab === 'user'">
				<div class="info-card animate-in">
					<div class="info-icon">
						<Sparkles :size="24" />
					</div>
					<div class="info-text">
						<h3>什么是 MCP？</h3>
						<p>Model Context Protocol (MCP) 是标准 AI 上下文协议。您可以申请 API Key，将您的记账数据以安全沙箱模式共享给 Cline、Cursor 等 AI 助手，实现 AI 一句话智能记账或图表语义检索。</p>
					</div>
				</div>

				<div class="list-section">
					<div class="list-header">
						<h2>我的 API 密钥</h2>
						<button 
							class="btn-apply" 
							@click="showApplyForm = !showApplyForm"
							v-if="!showApplyForm"
						>
							<Plus :size="16" /> 申请新密钥
						</button>
					</div>

					<!-- 申请表单 -->
					<div class="card apply-form animate-in" v-if="showApplyForm">
						<h3 class="form-title">申请 MCP API Key</h3>
						<div class="input-group">
							<label>申请应用名称 / 申请人</label>
							<input 
								class="input" 
								v-model="form.applicant" 
								placeholder="例如：我的 Cursor 助理 / Cline 浏览器扩展"
							/>
						</div>
						<div class="input-group">
							<label>使用用途说明</label>
							<textarea 
								class="textarea" 
								v-model="form.purpose" 
								placeholder="例如：在本地开发工具中接入记账系统，实现一句话快捷记账..."
								rows="3"
							></textarea>
						</div>
						<div class="form-actions">
							<button class="btn btn-secondary" @click="showApplyForm = false" :disabled="submitting">取消</button>
							<button class="btn btn-primary" @click="handleApply" :disabled="submitting">
								{{ submitting ? '提交中...' : '提交申请' }}
							</button>
						</div>
					</div>

					<!-- 数据列表 -->
					<div v-if="loading" class="state-hint">数据加载中...</div>
					<div v-else-if="myKeys.length === 0" class="state-hint">
						<Clock :size="32" style="opacity: 0.3; margin-bottom: 8px;" />
						<p>暂无 API 密钥，点击右上角开始申请</p>
					</div>
					<div v-else class="keys-grid">
						<div class="key-card card" v-for="k in myKeys" :key="k.ID">
							<div class="card-header">
								<div>
									<h3 class="applicant-name">{{ k.applicant }}</h3>
									<span class="created-time">{{ new Date(k.CreatedAt).toLocaleDateString() }} 申请</span>
								</div>
								<!-- 状态标签 -->
								<div class="status-badge" :class="k.status">
									<CheckCircle v-if="k.status === 'approved'" :size="12" />
									<Clock v-else-if="k.status === 'pending'" :size="12" />
									<AlertCircle v-else-if="k.status === 'revoked'" :size="12" />
									<XCircle v-else :size="12" />
									<span>{{ k.status === 'approved' ? '已批准' : k.status === 'pending' ? '待审核' : k.status === 'revoked' ? '已吊销' : '已拒绝' }}</span>
								</div>
							</div>

							<p class="purpose-desc"><strong>申请说明：</strong>{{ k.purpose }}</p>

							<!-- 已批准状态：展示 Key 和配置说明 -->
							<div class="approved-content" v-if="k.status === 'approved'">
								<div class="config-block">
									<div class="config-label">
										<span>MCP 连接地址 (SSE)</span>
										<button class="btn-copy" @click="copyText(getMcpUrl(k.key))">
											<Copy :size="14" /> 复制地址
										</button>
									</div>
									<div class="code-box">
										<code>{{ getMcpUrl(k.key) }}</code>
									</div>
								</div>

								<div class="config-block">
									<div class="config-label">
										<span>API Key (密钥令牌)</span>
										<button class="btn-copy" @click="copyText(k.key)">
											<Copy :size="14" /> 复制密钥
										</button>
									</div>
									<div class="code-box">
										<code>{{ k.key }}</code>
									</div>
								</div>

								<div class="integration-guide">
									<strong>💡 快速接入指南：</strong>
									<p>在您的 Cursor 或 Cline 配置中，添加远程 SSE MCP 服务：</p>
									<ul>
										<li>地址：复制上方的连接地址</li>
										<li>认证头：<code>Authorization: Bearer {{ k.key.substring(0, 12) }}...</code></li>
									</ul>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Tab 2: 管理员 - 审批管理 -->
			<div v-else>
				<div class="list-section">
					<div class="list-header">
						<h2>全站审批列表</h2>
					</div>

					<div v-if="loading" class="state-hint">加载申请中...</div>
					<div v-else-if="adminKeys.length === 0" class="state-hint">
						<CheckCircle :size="32" style="color: var(--success); margin-bottom: 8px;" />
						<p>目前没有申请审批的记录</p>
					</div>
					<div v-else class="keys-grid">
						<div class="key-card card admin-card" v-for="k in adminKeys" :key="k.ID">
							<div class="card-header">
								<div>
									<h3 class="applicant-name">{{ k.applicant }}</h3>
									<span class="user-id">租户用户ID: {{ k.user_id.substring(0, 8) }}...</span>
								</div>
								<div class="status-badge" :class="k.status">
									<span>{{ k.status === 'approved' ? '已批准' : k.status === 'pending' ? '待审核' : k.status === 'revoked' ? '已吊销' : '已拒绝' }}</span>
								</div>
							</div>

							<p class="purpose-desc"><strong>用途：</strong>{{ k.purpose }}</p>

							<!-- 管理员操作动作 -->
							<div class="admin-actions">
								<!-- 待审批状态 -->
								<template v-if="k.status === 'pending'">
									<button 
										class="btn btn-secondary btn-sm reject" 
										@click="handleReject(k.ID)"
										:disabled="actionLoading === k.ID"
									>
										拒绝
									</button>
									<button 
										class="btn btn-primary btn-sm approve" 
										@click="handleApprove(k.ID)"
										:disabled="actionLoading === k.ID"
									>
										批准放行
									</button>
								</template>
								<!-- 已审批状态，支持随时吊销 -->
								<template v-if="k.status === 'approved'">
									<div class="approved-admin-info">
										<span>密钥已分发</span>
									</div>
									<button 
										class="btn btn-danger btn-sm" 
										@click="handleRevoke(k.ID)"
										:disabled="actionLoading === k.ID"
									>
										吊销密钥
									</button>
								</template>
								<!-- 其他已处理完状态 -->
								<template v-if="k.status === 'rejected' || k.status === 'revoked'">
									<span class="handled-text">该单已处理完毕</span>
								</template>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<style scoped>
.mcp-page {
	min-height: 100vh;
	background-color: var(--bg-base);
}

.page-content {
	padding: 16px;
	padding-top: calc(env(safe-area-inset-top) + 70px);
	padding-bottom: calc(env(safe-area-inset-bottom) + 30px);
	max-width: 600px;
	margin: 0 auto;
}

.tab-header {
	display: flex;
	background: var(--bg-card);
	border: 1px solid var(--border);
	border-radius: 10px;
	padding: 4px;
	margin-bottom: 20px;
}

.tab-btn {
	flex: 1;
	border: none;
	background: transparent;
	padding: 10px 0;
	font-size: 0.95rem;
	font-weight: 600;
	color: var(--text-secondary);
	cursor: pointer;
	border-radius: 8px;
	transition: all 0.2s;
}

.tab-btn.active {
	background: var(--primary);
	color: white;
}

.info-card {
	display: flex;
	gap: 16px;
	background: var(--primary-soft);
	border: 1.5px solid var(--primary-light);
	padding: 18px;
	border-radius: 16px;
	margin-bottom: 24px;
	color: var(--primary-dark);
}

.info-icon {
	flex-shrink: 0;
	width: 44px;
	height: 44px;
	border-radius: 50%;
	background: rgba(37, 99, 235, 0.12);
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--primary);
}

.info-text h3 {
	margin: 0 0 6px 0;
	font-size: 1rem;
	font-weight: 700;
}

.info-text p {
	margin: 0;
	font-size: 0.85rem;
	line-height: 1.5;
}

.list-section {
	display: flex;
	flex-direction: column;
	gap: 16px;
}

.list-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.list-header h2 {
	font-size: 1.1rem;
	color: var(--text-primary);
	margin: 0;
	font-weight: 700;
}

.btn-apply {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	background: var(--primary);
	color: white;
	border: none;
	padding: 8px 14px;
	border-radius: 8px;
	font-size: 0.85rem;
	font-weight: 600;
	cursor: pointer;
}

.apply-form {
	padding: 20px;
	border-radius: 16px;
	display: flex;
	flex-direction: column;
	gap: 16px;
	border: 1px solid var(--border);
	background: var(--bg-card);
}

.form-title {
	margin: 0;
	font-size: 1rem;
	font-weight: 700;
	color: var(--text-primary);
}

.textarea {
	padding: 12px;
	border: 1px solid var(--border);
	border-radius: 8px;
	background: var(--bg-card);
	color: var(--text-primary);
	font-size: 0.95rem;
	font-family: inherit;
	outline: none;
	resize: none;
}

.textarea:focus {
	border-color: var(--primary);
}

.form-actions {
	display: flex;
	justify-content: flex-end;
	gap: 10px;
}

.state-hint {
	text-align: center;
	padding: 48px 20px;
	color: var(--text-tertiary);
	font-size: 0.9rem;
}

.keys-grid {
	display: flex;
	flex-direction: column;
	gap: 16px;
}

.key-card {
	padding: 20px;
	background: var(--bg-card);
	border: 1px solid var(--border);
	border-radius: 16px;
}

.card-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	margin-bottom: 12px;
}

.applicant-name {
	margin: 0 0 4px 0;
	font-size: 1.05rem;
	font-weight: 700;
	color: var(--text-primary);
}

.created-time, .user-id {
	font-size: 0.75rem;
	color: var(--text-tertiary);
}

.status-badge {
	display: inline-flex;
	align-items: center;
	gap: 4px;
	padding: 4px 10px;
	border-radius: 20px;
	font-size: 0.75rem;
	font-weight: 600;
}

.status-badge.approved {
	background: var(--success-soft);
	color: var(--success);
}

.status-badge.pending {
	background: var(--primary-soft);
	color: var(--primary);
}

.status-badge.rejected {
	background: var(--danger-soft);
	color: var(--danger);
}

.status-badge.revoked {
	background: #f1f5f9;
	color: #64748b;
}

.purpose-desc {
	margin: 0 0 16px 0;
	font-size: 0.85rem;
	color: var(--text-secondary);
	line-height: 1.5;
}

.approved-content {
	border-top: 1px dashed var(--border);
	padding-top: 16px;
	display: flex;
	flex-direction: column;
	gap: 14px;
}

.config-block {
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.config-label {
	display: flex;
	justify-content: space-between;
	align-items: center;
	font-size: 0.8rem;
	font-weight: 600;
	color: var(--text-secondary);
}

.btn-copy {
	background: transparent;
	border: none;
	color: var(--primary);
	font-size: 0.8rem;
	font-weight: 600;
	cursor: pointer;
	display: inline-flex;
	align-items: center;
	gap: 4px;
}

.code-box {
	background: var(--bg-base);
	border-radius: 8px;
	padding: 10px 12px;
	overflow-x: auto;
}

.code-box code {
	font-family: monospace;
	font-size: 0.8rem;
	color: var(--text-primary);
	white-space: nowrap;
}

.integration-guide {
	background: #fffbeb;
	border: 1px solid #fef3c7;
	padding: 12px 14px;
	border-radius: 8px;
	color: #b45309;
	font-size: 0.8rem;
}

.integration-guide strong {
	display: block;
	margin-bottom: 4px;
}

.integration-guide p {
	margin: 0 0 6px 0;
}

.integration-guide ul {
	margin: 0;
	padding-left: 16px;
}

.integration-guide li {
	margin-bottom: 2px;
}

.admin-actions {
	border-top: 1px dashed var(--border);
	padding-top: 16px;
	display: flex;
	justify-content: flex-end;
	gap: 10px;
	align-items: center;
}

.btn-sm {
	padding: 8px 16px;
	font-size: 0.85rem;
	border-radius: 8px;
}

.btn-sm.reject {
	background: var(--danger-soft);
	color: var(--danger);
	border: none;
}

.btn-sm.approve {
	background: var(--success);
	color: white;
	border: none;
}

.approved-admin-info {
	flex: 1;
	font-size: 0.85rem;
	color: var(--success);
	font-weight: 600;
}

.handled-text {
	font-size: 0.85rem;
	color: var(--text-tertiary);
}
</style>
