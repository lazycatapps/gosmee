// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Layout,
  Typography,
  Space,
  Button,
  Dropdown,
  Input,
  Table,
  Tag,
  Modal,
  Form,
  Select,
  Switch,
  InputNumber,
  Collapse,
  Tabs,
  Card,
  Statistic,
  Descriptions,
  Tooltip,
  Alert,
  Empty,
  Spin,
  Row,
  Col,
  App as AntApp,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  CaretRightOutlined,
  StopOutlined,
  SyncOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  ThunderboltOutlined,
  CopyOutlined,
  LogoutOutlined,
  LoginOutlined,
  WarningOutlined,
  PauseCircleOutlined,
  FileSearchOutlined,
  InfoCircleOutlined,
  LinkOutlined,
} from '@ant-design/icons';
import packageJson from '../package.json';
import './App.css';

const APP_VERSION = process.env.REACT_APP_VERSION || packageJson.version;
const GIT_COMMIT = process.env.REACT_APP_GIT_COMMIT || 'dev';
const GIT_COMMIT_FULL = process.env.REACT_APP_GIT_COMMIT_FULL || 'development';
const GIT_BRANCH = process.env.REACT_APP_GIT_BRANCH || 'local';
const BUILD_TIME = process.env.REACT_APP_BUILD_TIME || 'dev-build';

const BACKEND_API_URL = window._env_?.BACKEND_API_URL || '';
const API_PREFIX = (BACKEND_API_URL || '').replace(/\/$/, '');
const MAX_LOG_LINES = 500;

const CLIENT_STATUS_MAP = {
  running: { color: 'success', label: '运行中' },
  stopped: { color: 'default', label: '已停止' },
  error: { color: 'error', label: '错误' },
};

const EVENT_IGNORE_OPTIONS = [
  { label: 'push', value: 'push' },
  { label: 'pull_request', value: 'pull_request' },
  { label: 'issue_comment', value: 'issue_comment' },
  { label: 'release', value: 'release' },
  { label: 'workflow_run', value: 'workflow_run' },
  { label: 'tag_push', value: 'tag_push' },
];

const buildApiUrl = (path) => `${API_PREFIX}${path}`;

const apiFetch = (path, options = {}) => {
  const url = buildApiUrl(path);
  const config = {
    credentials: 'include',
    ...options,
  };

  if (config.body && !(config.headers && config.headers['Content-Type'])) {
    config.headers = {
      ...(config.headers || {}),
      'Content-Type': 'application/json',
    };
  }

  return fetch(url, config);
};

const formatDateTime = (dateString) => {
  if (!dateString) {
    return '-';
  }
  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) {
    return dateString;
  }
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
};

const formatDuration = (seconds) => {
  if (!seconds || seconds <= 0) {
    return '-';
  }

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  const parts = [];
  if (days) {
    parts.push(`${days}天`);
  }
  if (hours) {
    parts.push(`${hours}小时`);
  }
  if (minutes) {
    parts.push(`${minutes}分钟`);
  }
  if (!parts.length || secs) {
    parts.push(`${secs}秒`);
  }

  return parts.slice(0, 3).join('');
};

const formatBytes = (value) => {
  if (value === null || value === undefined) {
    return '-';
  }
  if (value === 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  let bytes = value;
  let unitIndex = 0;

  while (bytes >= 1024 && unitIndex < units.length - 1) {
    bytes /= 1024;
    unitIndex += 1;
  }

  const formatted = bytes >= 10 || unitIndex === 0 ? Math.round(bytes) : bytes.toFixed(1);
  return `${formatted} ${units[unitIndex]}`;
};

const safeJsonParse = (value) => {
  if (!value) {
    return null;
  }
  try {
    return JSON.parse(value);
  } catch (error) {
    return null;
  }
};

const renderClientStatusTag = (status) => {
  const info = CLIENT_STATUS_MAP[status] || { color: 'default', label: status };
  return <Tag color={info.color}>{info.label}</Tag>;
};

const copyToClipboard = async (text, messageApi) => {
  try {
    await navigator.clipboard.writeText(text);
    if (messageApi?.success) {
      messageApi.success('已复制到剪贴板');
    }
  } catch (error) {
    if (messageApi?.error) {
      messageApi.error('复制失败，请手动复制');
    }
  }
};

function ClientFormModal({ open, onCancel, onSubmit, initialValues, loading }) {
  const [form] = Form.useForm();
  const isEditing = Boolean(initialValues?.id);

  useEffect(() => {
    if (!open) {
      return;
    }
    form.resetFields();
    form.setFieldsValue({
      name: initialValues?.name || '',
      description: initialValues?.description || '',
      smeeUrl: initialValues?.smeeUrl || '',
      targetUrl: initialValues?.targetUrl || '',
      targetTimeout: initialValues?.targetTimeout || 60,
      httpie: initialValues?.httpie ?? false,
      ignoreEvents: initialValues?.ignoreEvents || [],
      noReplay: initialValues?.noReplay ?? false,
      sseBufferSize: initialValues?.sseBufferSize || 1048576,
    });
  }, [open, initialValues, form]);

  const handleSubmit = () => {
    form
      .validateFields()
      .then((values) => {
        const payload = {
          name: values.name.trim(),
          description: values.description?.trim() || '',
          smeeUrl: values.smeeUrl.trim(),
          targetUrl: values.targetUrl.trim(),
          targetTimeout: values.targetTimeout || 60,
          httpie: values.httpie,
          ignoreEvents: values.ignoreEvents || [],
          noReplay: values.noReplay,
          sseBufferSize: values.sseBufferSize || 1048576,
        };
        onSubmit(payload);
      })
      .catch(() => {});
  };

  return (
    <Modal
      title={isEditing ? '编辑实例' : '创建实例'}
      open={open}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      destroyOnHidden
      width={680}
      okText={isEditing ? '保存修改' : '创建实例'}
      cancelText="取消"
    >
      <Form layout="vertical" form={form}>
        <Form.Item
          label="实例名称"
          name="name"
          rules={[
            { required: true, message: '请输入实例名称' },
            { max: 50, message: '实例名称不能超过 50 个字符' },
          ]}
        >
          <Input placeholder="例如：Agola Webhook 中继" allowClear />
        </Form.Item>

        <Form.Item
          label="描述"
          name="description"
          rules={[{ max: 200, message: '描述不能超过 200 个字符' }]}
        >
          <Input.TextArea
            rows={3}
            showCount
            maxLength={200}
            placeholder="可选：说明此实例的用途，便于团队成员识别"
          />
        </Form.Item>

        <Form.Item
          label="Smee URL"
          name="smeeUrl"
          rules={[
            { required: true, message: '请输入 Smee URL' },
            {
              validator: (_, value) => {
                if (!value) {
                  return Promise.resolve();
                }
                if (!/^https:\/\/.+/.test(value)) {
                  return Promise.reject(new Error('Smee URL 必须使用 HTTPS'));
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input
            placeholder="https://hook.pipelinesascode.com/xxxxxxxx"
            disabled={isEditing}
            allowClear
          />
        </Form.Item>

        <Form.Item
          label="Target URL"
          name="targetUrl"
          rules={[
            { required: true, message: '请输入 Target URL' },
            {
              validator: (_, value) => {
                if (!value) {
                  return Promise.resolve();
                }
                if (!/^https?:\/\/.+/.test(value)) {
                  return Promise.reject(new Error('请输入有效的 HTTP/HTTPS 地址'));
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input placeholder="https://internal.example.com/webhooks" allowClear />
        </Form.Item>

        <Collapse
          items={[
            {
              key: 'advanced',
              label: '高级选项',
              children: (
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                  <Form.Item
                    label="目标连接超时时间（秒）"
                    name="targetTimeout"
                    rules={[{ type: 'number', min: 1, message: '超时时间必须大于 0' }]}
                  >
                    <InputNumber min={1} max={600} style={{ width: '100%' }} />
                  </Form.Item>

                  <Form.Item label="脚本格式" name="httpie" valuePropName="checked">
                    <Switch checkedChildren="HTTPie" unCheckedChildren="cURL" />
                  </Form.Item>

                  <Form.Item label="忽略事件类型" name="ignoreEvents">
                    <Select
                      mode="multiple"
                      allowClear
                      placeholder="选择需要忽略的事件类型"
                      options={EVENT_IGNORE_OPTIONS}
                    />
                  </Form.Item>

                  <Form.Item label="仅保存不转发" name="noReplay" valuePropName="checked">
                    <Switch checkedChildren="仅保存" unCheckedChildren="正常转发" />
                  </Form.Item>

                  <Form.Item
                    label="SSE 缓冲区大小（字节）"
                    name="sseBufferSize"
                    rules={[{ type: 'number', min: 1024, message: '缓冲区至少 1024 字节' }]}
                  >
                    <InputNumber min={1024} step={1024} style={{ width: '100%' }} />
                  </Form.Item>
                </Space>
              ),
            },
          ]}
        />
      </Form>
    </Modal>
  );
}

function EventDetailModal({ open, loading, event, onClose, onReplay, onDelete, messageApi }) {
  const payloadJson = useMemo(() => safeJsonParse(event?.payload), [event]);
  const headersArray = useMemo(() => {
    if (!event?.headers) {
      return [];
    }
    return Object.entries(event.headers).map(([key, value]) => ({ key, value }));
  }, [event]);

  const responseJson = useMemo(() => safeJsonParse(event?.response), [event]);

  const tabItems = useMemo(() => {
    const items = [
      {
        key: 'payload',
        label: 'Payload',
        children: (
          <Card size="small" className="code-block">
            <Space style={{ marginBottom: 8 }}>
              <Button
                icon={<CopyOutlined />}
                size="small"
                onClick={() => event && copyToClipboard(event.payload || '', messageApi)}
              >
                复制
              </Button>
            </Space>
            <pre>
              {payloadJson
                ? JSON.stringify(payloadJson, null, 2)
                : event?.payload || '无数据'}
            </pre>
          </Card>
        ),
      },
    ];

    if (headersArray.length > 0) {
      items.push({
        key: 'headers',
        label: '请求头',
        children: (
          <Card size="small">
            <Table
              pagination={false}
              size="small"
              dataSource={headersArray}
              columns={[
                { title: 'Header', dataIndex: 'key', key: 'key', width: '40%' },
                { title: '值', dataIndex: 'value', key: 'value' },
              ]}
              locale={{ emptyText: '暂无请求头信息' }}
              rowKey="key"
            />
          </Card>
        ),
      });
    }

    if (event?.response) {
      items.push({
        key: 'response',
        label: '响应内容',
        children: (
          <Card size="small" className="code-block">
            <pre>
              {responseJson
                ? JSON.stringify(responseJson, null, 2)
                : event?.response || '无响应数据'}
            </pre>
          </Card>
        ),
      });
    }

    return items;
  }, [event, payloadJson, headersArray, responseJson, messageApi]);

  return (
    <Modal
      title="事件详情"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          关闭
        </Button>,
        <Button
          key="replay"
          type="primary"
          icon={<ThunderboltOutlined />}
          onClick={() => event && onReplay(event.id)}
          disabled={!event}
        >
          重新转发
        </Button>,
        <Button
          key="delete"
          danger
          icon={<DeleteOutlined />}
          onClick={() => event && onDelete(event.id)}
          disabled={!event}
        >
          删除
        </Button>,
      ]}
      width={900}
      destroyOnHidden
    >
      {loading && (
        <div style={{ textAlign: 'center', padding: 24 }}>
          <Spin />
        </div>
      )}

      {!loading && !event && <Empty description="未找到事件详情" />}

      {!loading && event && (
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="事件 ID" span={2}>
              <Typography.Text copyable>{event.id}</Typography.Text>
            </Descriptions.Item>
            <Descriptions.Item label="时间戳" span={2}>
              {formatDateTime(event.timestamp)}
            </Descriptions.Item>
          </Descriptions>

          <Tabs
            defaultActiveKey="payload"
            items={tabItems}
          />
        </Space>
      )}
    </Modal>
  );
}

function AppContent() {
  const { message, modal } = AntApp.useApp();

  const [authInfo, setAuthInfo] = useState({ loading: true, data: null, error: null });
  const [quotaInfo, setQuotaInfo] = useState({ loading: true, quota: null, warning: '' });

  const [clients, setClients] = useState([]);
  const [clientsTotal, setClientsTotal] = useState(0);
  const [clientsLoading, setClientsLoading] = useState(false);
  const [clientsPage, setClientsPage] = useState(1);
  const [clientsPageSize, setClientsPageSize] = useState(20);
  const [selectedClientKeys, setSelectedClientKeys] = useState([]);
  const [clientBatchLoading, setClientBatchLoading] = useState({
    start: false,
    stop: false,
    startAll: false,
    stopAll: false,
  });
  const [statusFilter, setStatusFilter] = useState('all');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortState, setSortState] = useState({ field: 'createdAt', order: 'desc' });

  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [formSubmitting, setFormSubmitting] = useState(false);
  const [editingClient, setEditingClient] = useState(null);

  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedClientId, setSelectedClientId] = useState(null);
  const [clientDetail, setClientDetail] = useState(null);
  const [clientStats, setClientStats] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailTab, setDetailTab] = useState('overview');

  const [events, setEvents] = useState([]);
  const [eventPagination, setEventPagination] = useState({ page: 1, pageSize: 10, total: 0 });
  const [eventSortState, setEventSortState] = useState({ field: 'timestamp', order: 'desc' });
  const [eventsLoading, setEventsLoading] = useState(false);
  const [selectedEventKeys, setSelectedEventKeys] = useState([]);
  const [eventDetailVisible, setEventDetailVisible] = useState(false);
  const [eventDetailLoading, setEventDetailLoading] = useState(false);
  const [eventDetailData, setEventDetailData] = useState(null);

  const [logs, setLogs] = useState([]);
  const [logsAutoScroll, setLogsAutoScroll] = useState(true);
  const logsContainerRef = useRef(null);
  const logEventSourceRef = useRef(null);

  const fetchAuthInfo = useCallback(async () => {
    try {
      const response = await apiFetch('/api/v1/auth/userinfo');
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(data.error || '无法获取用户信息');
      }
      setAuthInfo({ loading: false, data, error: null });
    } catch (error) {
      setAuthInfo({
        loading: false,
        data: { authenticated: false, oidc_enabled: false },
        error: error.message,
      });
    }
  }, []);

  const fetchQuotaInfo = useCallback(async () => {
    try {
      const response = await apiFetch('/api/v1/quota');
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        // If it's a 401 Unauthorized, don't show error message
        // The user should see the login prompt instead
        if (response.status === 401) {
          setQuotaInfo({ loading: false, quota: null, warning: '' });
          return;
        }
        throw new Error(data.error || '无法获取配额信息');
      }
      setQuotaInfo({
        loading: false,
        quota: data.quota,
        warning: data.warning || '',
      });
    } catch (error) {
      setQuotaInfo({ loading: false, quota: null, warning: '' });
      // Only show error message if it's not a network error during unauthenticated state
      if (error.message !== 'Failed to fetch') {
        message.error(`加载配额信息失败：${error.message}`);
      }
    }
  }, [message]);

  const loadClients = useCallback(async () => {
    setClientsLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(clientsPage),
        pageSize: String(clientsPageSize),
        sortBy: sortState.field,
        sortOrder: sortState.order,
      });
      if (statusFilter !== 'all') {
        params.set('status', statusFilter);
      }
      if (searchKeyword.trim()) {
        params.set('search', searchKeyword.trim());
      }

      const response = await apiFetch(`/api/v1/clients?${params.toString()}`);
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        // If it's a 401 Unauthorized, don't show error message
        // The user should see the login prompt instead
        if (response.status === 401) {
          setClients([]);
          setClientsTotal(0);
          return;
        }
        throw new Error(data.error || '加载实例列表失败');
      }

      const clientList = data.clients || [];
      setClients(clientList);
      setClientsTotal(data.total || 0);
      setSelectedClientKeys((prev) =>
        prev.filter((key) => clientList.some((client) => client.id === key)),
      );
    } catch (error) {
      setClients([]);
      setClientsTotal(0);
      // Only show error message if it's not a network error during unauthenticated state
      if (error.message !== 'Failed to fetch') {
        message.error(`加载实例列表失败：${error.message}`);
      }
    } finally {
      setClientsLoading(false);
    }
  }, [clientsPage, clientsPageSize, sortState, statusFilter, searchKeyword, message, setSelectedClientKeys]);

  const requestClientDetail = useCallback(async (clientId) => {
    const response = await apiFetch(`/api/v1/clients/${clientId}`);
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || '获取实例详情失败');
    }
    return data;
  }, []);

  const requestClientStats = useCallback(async (clientId) => {
    const response = await apiFetch(`/api/v1/clients/${clientId}/stats`);
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || '获取实例统计信息失败');
    }
    return data;
  }, []);

  const fetchClientDetail = useCallback(
    async (clientId) => {
      setDetailLoading(true);
      try {
        const detail = await requestClientDetail(clientId);
        setClientDetail(detail);
        try {
          const stats = await requestClientStats(clientId);
          setClientStats(stats);
        } catch (statsError) {
          message.warning(`加载实例统计信息失败：${statsError.message}`);
          setClientStats(null);
        }
      } catch (error) {
        message.error(`加载实例详情失败：${error.message}`);
      } finally {
        setDetailLoading(false);
      }
    },
    [message, requestClientDetail, requestClientStats],
  );

  const loadEvents = useCallback(
    async (clientId, page = eventPagination.page, pageSize = eventPagination.pageSize) => {
      setEventsLoading(true);
      try {
        const params = new URLSearchParams({
          page: String(page),
          pageSize: String(pageSize),
          sortBy: eventSortState.field,
          sortOrder: eventSortState.order,
        });

        const response = await apiFetch(
          `/api/v1/clients/${clientId}/events?${params.toString()}`,
        );
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data.error || '加载事件失败');
        }

        setEvents(data.events || []);
        setEventPagination({
          page: data.page || page,
          pageSize: data.pageSize || pageSize,
          total: data.total || 0,
        });
      } catch (error) {
        message.error(`加载事件失败：${error.message}`);
      } finally {
        setEventsLoading(false);
      }
    },
    [eventPagination.page, eventPagination.pageSize, eventSortState, message],
  );

  const handleRefreshDetail = useCallback(() => {
    if (!clientDetail?.id) return;

    // 刷新实例详情
    fetchClientDetail(clientDetail.id);

    // 如果当前在事件标签页,也刷新事件列表
    if (detailTab === 'events') {
      loadEvents(clientDetail.id);
    }
  }, [clientDetail, detailTab, fetchClientDetail, loadEvents]);

  useEffect(() => {
    fetchAuthInfo();
  }, [fetchAuthInfo]);

  useEffect(() => {
    // Only fetch quota and clients if auth info is loaded
    if (authInfo.loading) {
      return;
    }
    // If OIDC is enabled and user is not authenticated, don't make these API calls
    if (authInfo.data?.oidc_enabled && !authInfo.data?.authenticated) {
      return;
    }
    fetchQuotaInfo();
    loadClients();
  }, [authInfo.loading, authInfo.data?.oidc_enabled, authInfo.data?.authenticated, fetchQuotaInfo, loadClients]);

  useEffect(() => {
    if (!detailVisible || !selectedClientId || detailTab !== 'events') {
      return;
    }
    loadEvents(selectedClientId);
  }, [detailVisible, selectedClientId, detailTab, loadEvents]);

  useEffect(() => {
    if (!detailVisible || !selectedClientId || detailTab !== 'logs') {
      if (logEventSourceRef.current) {
        logEventSourceRef.current.close();
        logEventSourceRef.current = null;
      }
      return;
    }

    setLogs([]);
    if (logEventSourceRef.current) {
      logEventSourceRef.current.close();
      logEventSourceRef.current = null;
    }

    try {
      const source = new EventSource(
        buildApiUrl(`/api/v1/clients/${selectedClientId}/logs/stream`),
        { withCredentials: true },
      );
      logEventSourceRef.current = source;
      source.addEventListener('log', (event) => {
        setLogs((prev) => {
          const next = [...prev, event.data];
          if (next.length > MAX_LOG_LINES) {
            next.splice(0, next.length - MAX_LOG_LINES);
          }
          return next;
        });
      });
      source.onerror = () => {
        message.warning('日志流连接异常，稍后将自动重试');
      };
    } catch (error) {
      message.error(`建立日志流失败：${error.message}`);
    }

    return () => {
      if (logEventSourceRef.current) {
        logEventSourceRef.current.close();
        logEventSourceRef.current = null;
      }
    };
  }, [detailVisible, selectedClientId, detailTab, message]);

  useEffect(() => {
    if (!logsAutoScroll) {
      return;
    }
    if (logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [logs, logsAutoScroll]);

  const handleOpenCreate = () => {
    setEditingClient(null);
    setCreateModalVisible(true);
  };

  const handleOpenEdit = useCallback(
    async (clientId) => {
      try {
        const detail = await requestClientDetail(clientId);
        setEditingClient(detail);
        setCreateModalVisible(true);
      } catch (error) {
        message.error(`加载实例信息失败：${error.message}`);
      }
    },
    [message, requestClientDetail],
  );

  const handleSubmitClient = async (payload) => {
    setFormSubmitting(true);
    try {
      const method = editingClient ? 'PUT' : 'POST';
      const url = editingClient ? `/api/v1/clients/${editingClient.id}` : '/api/v1/clients';
      const response = await apiFetch(url, {
        method,
        body: JSON.stringify(payload),
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(data.error || '保存实例失败');
      }

      message.success(editingClient ? '实例已更新' : '实例创建成功');
      setCreateModalVisible(false);
      setEditingClient(null);
      if (!editingClient) {
        setClientsPage(1);
      }
      await loadClients();
      if (editingClient && selectedClientId === editingClient.id) {
        fetchClientDetail(editingClient.id);
      }
      fetchQuotaInfo();
    } catch (error) {
      message.error(`保存失败：${error.message}`);
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleDeleteClient = useCallback(
    (client) => {
      modal.confirm({
        title: `确认删除实例「${client.name}」吗？`,
        icon: <WarningOutlined />,
        content: '删除后，该实例的配置、日志与事件历史将无法恢复。',
        okText: '删除',
        okType: 'danger',
        cancelText: '取消',
        async onOk() {
          try {
            const response = await apiFetch(`/api/v1/clients/${client.id}`, {
              method: 'DELETE',
            });
            const data = await response.json().catch(() => ({}));
            if (!response.ok) {
              throw new Error(data.error || '删除实例失败');
            }
            message.success('实例已删除');
            if (selectedClientId === client.id) {
              setDetailVisible(false);
              setSelectedClientId(null);
              setClientDetail(null);
              setClientStats(null);
            }
            await loadClients();
            setSelectedClientKeys((prev) => prev.filter((key) => key !== client.id));
            fetchQuotaInfo();
          } catch (error) {
            message.error(`删除失败：${error.message}`);
          }
        },
      });
    },
    [
      fetchQuotaInfo,
      loadClients,
      message,
      modal,
      selectedClientId,
      setSelectedClientKeys,
    ],
  );

  const performClientAction = useCallback(
    async (clientId, action) => {
      const actionLabelMap = {
        start: '启动',
        stop: '停止',
        restart: '重启',
      };
      try {
        const response = await apiFetch(`/api/v1/clients/${clientId}/${action}`, {
          method: 'POST',
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data.error || '操作失败');
        }
        message.success(`${actionLabelMap[action]}指令已发送`);
        await loadClients();
        if (selectedClientId === clientId) {
          fetchClientDetail(clientId);
        }
      } catch (error) {
        message.error(`${actionLabelMap[action]}实例失败：${error.message}`);
      }
    },
    [fetchClientDetail, loadClients, message, selectedClientId],
  );

  const handleBatchClientAction = useCallback(
    async (action, options = {}) => {
      const { all = false } = options;
      const loadingKey = all ? `${action}All` : action;

      if (!all && selectedClientKeys.length === 0) {
        message.warning('请选择需要操作的实例');
        return;
      }

      setClientBatchLoading((prev) => ({ ...prev, [loadingKey]: true }));

      try {
        const payload = all ? { all: true } : { clientIds: selectedClientKeys };
        const response = await apiFetch(`/api/v1/clients/batch/${action}`, {
          method: 'POST',
          body: JSON.stringify(payload),
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data.error || `${action === 'start' ? '启动' : '停止'}实例失败`);
        }

        const total = data.total || 0;
        const successful = data.successful || 0;
        const failed = data.failed || 0;

        if (total === 0) {
          message.info('没有可操作的实例');
        } else {
          message.success(
            `${action === 'start' ? '启动' : '停止'}完成：成功 ${successful} 个，失败 ${failed} 个`,
          );
        }

        if (!all) {
          setSelectedClientKeys([]);
        }

        await loadClients();
        if (detailVisible && selectedClientId) {
          await fetchClientDetail(selectedClientId);
        }
      } catch (error) {
        message.error(error.message || `${action === 'start' ? '启动' : '停止'}操作失败`);
      } finally {
        setClientBatchLoading((prev) => ({ ...prev, [loadingKey]: false }));
      }
    },
    [detailVisible, fetchClientDetail, loadClients, message, selectedClientId, selectedClientKeys],
  );

  const handleViewClient = useCallback(
    (clientId) => {
      setSelectedClientId(clientId);
      setDetailVisible(true);
      setDetailTab('overview');
      fetchClientDetail(clientId);
    },
    [fetchClientDetail],
  );

  const handleDetailClose = () => {
    setDetailVisible(false);
    setSelectedClientId(null);
    setClientDetail(null);
    setClientStats(null);
    setEvents([]);
    setSelectedEventKeys([]);
    setEventPagination({ page: 1, pageSize: 10, total: 0 });
    setLogs([]);
    if (logEventSourceRef.current) {
      logEventSourceRef.current.close();
      logEventSourceRef.current = null;
    }
  };

  const handleLogin = () => {
    window.location.href = buildApiUrl('/api/v1/auth/login');
  };

  const handleLogout = async () => {
    try {
      const response = await apiFetch('/api/v1/auth/logout', { method: 'POST' });
      if (!response.ok) {
        throw new Error('登出失败');
      }
      message.success('已退出登录');
      fetchAuthInfo();
    } catch (error) {
      message.error(`退出登录失败：${error.message}`);
    }
  };

  const handleReplayEvents = useCallback(
    async (eventIds) => {
      if (!selectedClientId) {
        return;
      }
      try {
        const response = await apiFetch(`/api/v1/clients/${selectedClientId}/events/replay`, {
          method: 'POST',
          body: JSON.stringify({ eventIds }),
        });
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data.error || '重放事件失败');
        }
        message.success(`重放完成：成功 ${data.successful} 条，失败 ${data.failed} 条`);
        loadEvents(selectedClientId);
      } catch (error) {
        message.error(`重放失败：${error.message}`);
      }
    },
    [loadEvents, message, selectedClientId],
  );

  const handleDeleteEvent = useCallback(
    (eventId) => {
      if (!selectedClientId) {
        return;
      }
      modal.confirm({
        title: '确认删除该事件吗？',
        icon: <WarningOutlined />,
        okText: '删除',
        okType: 'danger',
        cancelText: '取消',
        async onOk() {
          try {
            const response = await apiFetch(
              `/api/v1/clients/${selectedClientId}/events/${eventId}`,
              { method: 'DELETE' },
            );
            const data = await response.json().catch(() => ({}));
            if (!response.ok) {
              throw new Error(data.error || '删除事件失败');
            }
            message.success('事件已删除');
            loadEvents(selectedClientId);
          } catch (error) {
            message.error(`删除失败：${error.message}`);
          }
        },
      });
    },
    [loadEvents, message, modal, selectedClientId],
  );

  const handleOpenEventDetail = useCallback(
    async (eventId) => {
      if (!selectedClientId) {
        return;
      }
      setEventDetailVisible(true);
      setEventDetailLoading(true);
      try {
        const response = await apiFetch(
          `/api/v1/clients/${selectedClientId}/events/${eventId}`,
        );
        const data = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(data.error || '获取事件详情失败');
        }
        setEventDetailData(data);
      } catch (error) {
        message.error(`加载事件详情失败：${error.message}`);
        setEventDetailVisible(false);
      } finally {
        setEventDetailLoading(false);
      }
    },
    [message, selectedClientId],
  );

  const clientColumns = useMemo(
    () => [
      {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
        sorter: true,
        render: (value, record) => (
          <Space direction="vertical" size={0}>
            <Typography.Text
              strong
              style={{ cursor: 'pointer' }}
              onClick={() => handleViewClient(record.id)}
            >
              {value}
            </Typography.Text>
            {record.description && (
              <Typography.Text type="secondary" ellipsis style={{ maxWidth: 220 }}>
                {record.description}
              </Typography.Text>
            )}
          </Space>
        ),
      },
      {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        sorter: true,
        render: (value) => renderClientStatusTag(value),
      },
      {
        title: 'Smee URL',
        dataIndex: 'smeeUrl',
        key: 'smeeUrl',
        render: (value) => (
          <Tooltip
            overlayInnerStyle={{ maxWidth: 520 }}
            title={
              value ? (
                <Space direction="vertical" size={0}>
                  <span className="tooltip-url">{value}</span>
                  <Typography.Text type="secondary">点击复制</Typography.Text>
                </Space>
              ) : (
                '无'
              )
            }
          >
            <Typography.Text
              ellipsis
              style={{ maxWidth: 240 }}
              className="copyable-text"
              onClick={() => copyToClipboard(value || '', message)}
            >
              {value}
            </Typography.Text>
          </Tooltip>
        ),
      },
      {
        title: 'Target URL',
        dataIndex: 'targetUrl',
        key: 'targetUrl',
        render: (value) => (
          <Tooltip
            overlayInnerStyle={{ maxWidth: 520 }}
            title={
              value ? (
                <Space direction="vertical" size={0}>
                  <span className="tooltip-url">{value}</span>
                  <Typography.Text type="secondary">点击复制</Typography.Text>
                </Space>
              ) : (
                '无'
              )
            }
          >
            <Typography.Text
              ellipsis
              style={{ maxWidth: 240 }}
              className="copyable-text"
              onClick={() => copyToClipboard(value || '', message)}
            >
              {value}
            </Typography.Text>
          </Tooltip>
        ),
      },
      {
        title: '最后活动时间',
        dataIndex: 'lastActivity',
        key: 'lastActivity',
        render: (value) => (value ? formatDateTime(value) : '-'),
      },
      {
        title: '操作',
        key: 'actions',
        width: 280,
        render: (_, record) => (
          <Space size="small" wrap>
            <Button
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewClient(record.id)}
            >
              详情
            </Button>
            {record.status === 'running' ? (
              <Button
                size="small"
                icon={<StopOutlined />}
                onClick={() => performClientAction(record.id, 'stop')}
              >
                停止
              </Button>
            ) : (
              <Button
                size="small"
                type="primary"
                icon={<CaretRightOutlined />}
                onClick={() => performClientAction(record.id, 'start')}
              >
                启动
              </Button>
            )}
            <Tooltip title="重启">
              <Button
                size="small"
                icon={<SyncOutlined />}
                onClick={() => performClientAction(record.id, 'restart')}
              />
            </Tooltip>
            <Tooltip title="编辑">
              <Button
                size="small"
                icon={<EditOutlined />}
                onClick={() => handleOpenEdit(record.id)}
              />
            </Tooltip>
            <Tooltip title="删除">
              <Button
                size="small"
                danger
                icon={<DeleteOutlined />}
                onClick={() => handleDeleteClient(record)}
              />
            </Tooltip>
          </Space>
        ),
      },
    ],
    [handleDeleteClient, handleOpenEdit, handleViewClient, message, performClientAction],
  );

  const eventColumns = useMemo(
    () => [
      {
        title: '时间',
        dataIndex: 'timestamp',
        key: 'timestamp',
        sorter: true,
        width: 200,
        render: (value) => formatDateTime(value),
      },
      {
        title: '事件 ID',
        dataIndex: 'id',
        key: 'id',
        ellipsis: true,
        render: (value) => (
          <Typography.Text copyable={{ text: value }} ellipsis style={{ maxWidth: '100%' }}>
            {value}
          </Typography.Text>
        ),
      },
      {
        title: '操作',
        key: 'actions',
        width: 240,
        fixed: 'right',
        render: (_, record) => (
          <Space size="small" wrap>
            <Button
              size="small"
              icon={<FileSearchOutlined />}
              onClick={() => handleOpenEventDetail(record.id)}
            >
              查看
            </Button>
            <Button
              size="small"
              icon={<ThunderboltOutlined />}
              onClick={() => handleReplayEvents([record.id])}
            >
              重放
            </Button>
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteEvent(record.id)}
            >
              删除
            </Button>
          </Space>
        ),
      },
    ],
    [handleDeleteEvent, handleOpenEventDetail, handleReplayEvents],
  );

  const statsGrid = clientStats ? (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={12} md={8} lg={6}>
        <Card size="small">
          <Statistic title="运行时长" value={formatDuration(clientStats.runningTime)} />
        </Card>
      </Col>
      <Col xs={24} sm={12} md={10} lg={8}>
        <Card size="small">
          <Statistic title="最后事件时间" value={formatDateTime(clientStats.lastEventTime)} />
        </Card>
      </Col>
    </Row>
  ) : (
    <Alert
      type="info"
      message="暂无统计数据"
      description="实例尚未产生事件或统计信息正在收集中。"
      showIcon
    />
  );

  const isAuthenticated = authInfo.data?.authenticated === true;
  const isOidcEnabled = authInfo.data?.oidc_enabled === true;

  return (
    <Layout className="app-layout">
      <Layout.Header className="app-header">
        <div className="app-header-inner">
          <Typography.Title level={4} style={{ margin: 0 }}>
            Gosmee Web UI
          </Typography.Title>
          {!authInfo.loading && isAuthenticated && (
            <div className="app-header-actions">
              <div className="app-header-user-info">
                <Typography.Text className="app-header-user-email" ellipsis={{ tooltip: authInfo.data.email || authInfo.data.user_id }}>
                  {authInfo.data.email || authInfo.data.user_id}
                </Typography.Text>
                <Button icon={<LogoutOutlined />} onClick={handleLogout}>
                  登出
                </Button>
              </div>
            </div>
          )}
        </div>
      </Layout.Header>

      <Layout.Content className="app-content">
        {authInfo.loading ? (
          <div className="app-content-inner">
            <div className="app-loading">
              <Spin size="large" />
            </div>
          </div>
        ) : isOidcEnabled && !isAuthenticated ? (
          <div className="app-content-inner">
            <div className="app-login-wrapper">
              <Card style={{ maxWidth: 500, textAlign: 'center', width: '100%' }}>
                <Space direction="vertical" size="large" style={{ width: '100%' }}>
                  <LoginOutlined style={{ fontSize: 64, color: '#1890ff' }} />
                  <Typography.Title level={3}>请登录后使用</Typography.Title>
                  <Typography.Paragraph type="secondary">
                    登录后您可以创建和管理 Webhook 转发实例,查看事件历史和统计信息。
                  </Typography.Paragraph>
                  <Button type="primary" size="large" icon={<LoginOutlined />} onClick={handleLogin}>
                    登录
                  </Button>
                </Space>
              </Card>
            </div>
          </div>
        ) : (
          <div className="app-content-inner">
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              {quotaInfo.warning && (
                <Alert
                  type="warning"
                  message="存储使用预警"
                  description={quotaInfo.warning}
                  showIcon
                  closable
                />
              )}
              <Card size="small" style={{ marginBottom: 16, width: '100%' }}>
                <Space direction="vertical" size="small" style={{ width: '100%' }}>
                  <Typography.Text type="secondary">
                    <InfoCircleOutlined style={{ marginRight: 8 }} />
                    常用 Webhook 中继服务（点击获取新的转发地址）:
                  </Typography.Text>
                  <Space wrap size="middle">
                    <Button
                      type="link"
                      icon={<LinkOutlined />}
                      href="https://hook.pipelinesascode.com/"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Pipelines as Code Hook
                    </Button>
                    <Button
                      type="link"
                      icon={<LinkOutlined />}
                      href="https://smee.io/"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Smee.io
                    </Button>
                  </Space>
                </Space>
              </Card>

          <Card
            title="实例列表"
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={() => loadClients()}>
                  刷新
                </Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={handleOpenCreate}>
                  创建实例
                </Button>
              </Space>
            }
            style={{ width: '100%' }}
          >
            <Space className="clients-toolbar" wrap align="center">
              <Input.Search
                placeholder="按名称搜索"
                allowClear
                onSearch={(value) => {
                  setClientsPage(1);
                  setSearchKeyword(value);
                }}
              />
              <Select
                value={statusFilter}
                options={[
                  { label: '全部状态', value: 'all' },
                  { label: '运行中', value: 'running' },
                  { label: '已停止', value: 'stopped' },
                  { label: '错误', value: 'error' },
                ]}
                onChange={(value) => {
                  setClientsPage(1);
                  setStatusFilter(value);
                }}
                style={{ width: 140 }}
              />
              <Dropdown.Button
                type="primary"
                icon={<CaretRightOutlined />}
                loading={
                  clientBatchLoading.start ||
                  clientBatchLoading.stop ||
                  clientBatchLoading.startAll ||
                  clientBatchLoading.stopAll
                }
                menu={{
                  items: [
                    {
                      key: 'startSelected',
                      label: '启动选中',
                      disabled: !selectedClientKeys.length,
                    },
                    {
                      key: 'stopSelected',
                      label: '停止选中',
                      disabled: !selectedClientKeys.length,
                      danger: true,
                    },
                    { type: 'divider' },
                    {
                      key: 'startAll',
                      label: '启动全部',
                    },
                    {
                      key: 'stopAll',
                      label: '停止全部',
                      danger: true,
                    },
                  ],
                  onClick: ({ key }) => {
                    switch (key) {
                      case 'startSelected':
                        handleBatchClientAction('start');
                        break;
                      case 'stopSelected':
                        handleBatchClientAction('stop');
                        break;
                      case 'startAll':
                        handleBatchClientAction('start', { all: true });
                        break;
                      case 'stopAll':
                        handleBatchClientAction('stop', { all: true });
                        break;
                      default:
                        break;
                    }
                  },
                }}
                onClick={() => handleBatchClientAction('start')}
              >
                批量操作
              </Dropdown.Button>
            </Space>

            <Table
              rowKey="id"
              dataSource={clients}
              columns={clientColumns}
              loading={clientsLoading}
              scroll={{ x: 960 }}
              rowSelection={{
                selectedRowKeys: selectedClientKeys,
                onChange: (keys) => setSelectedClientKeys(keys),
                preserveSelectedRowKeys: true,
              }}
              pagination={{
                current: clientsPage,
                pageSize: clientsPageSize,
                total: clientsTotal,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个实例`,
              }}
              onChange={(pagination, filters, sorter) => {
                setClientsPage(pagination.current);
                setClientsPageSize(pagination.pageSize);
                if (sorter && sorter.field) {
                  setSortState({
                    field: sorter.field,
                    order: sorter.order === 'ascend' ? 'asc' : 'desc',
                  });
                } else {
                  setSortState({ field: 'createdAt', order: 'desc' });
                }
              }}
            />
          </Card>
        </Space>
        </div>
        )}
      </Layout.Content>

      <Layout.Footer className="app-footer">
        <Typography.Text type="secondary">
          Gosmee Web UI · v{APP_VERSION} ·{' '}
          <Typography.Link
            href={`https://github.com/lazycatapps/gosmee/commit/${GIT_COMMIT_FULL}`}
            target="_blank"
            rel="noopener noreferrer"
          >
            {GIT_COMMIT}
          </Typography.Link>
          {' '}· Copyright © 2025 Lazycat Apps ·{' '}
          <Typography.Link
            href="https://github.com/lazycatapps/gosmee"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub
          </Typography.Link>
        </Typography.Text>
      </Layout.Footer>

      <ClientFormModal
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          setEditingClient(null);
        }}
        onSubmit={handleSubmitClient}
        initialValues={editingClient}
        loading={formSubmitting}
      />

      <Modal
        title={clientDetail ? `实例详情 · ${clientDetail.name}` : '实例详情'}
        open={detailVisible}
        onCancel={handleDetailClose}
        width={1000}
        footer={null}
        destroyOnHidden
        maskClosable={false}
      >
        {detailLoading && (
          <div style={{ textAlign: 'center', padding: 24 }}>
            <Spin />
          </div>
        )}

        {!detailLoading && !clientDetail && <Empty description="暂无详情数据" />}

        {!detailLoading && clientDetail && (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Space wrap>
              <Button
                type="primary"
                icon={<ReloadOutlined />}
                onClick={handleRefreshDetail}
              >
                刷新
              </Button>
              {clientDetail.status === 'running' ? (
                <Button
                  icon={<StopOutlined />}
                  onClick={() => performClientAction(clientDetail.id, 'stop')}
                >
                  停止
                </Button>
              ) : (
                <Button
                  type="primary"
                  icon={<CaretRightOutlined />}
                  onClick={() => performClientAction(clientDetail.id, 'start')}
                >
                  启动
                </Button>
              )}
              <Button
                icon={<SyncOutlined />}
                onClick={() => performClientAction(clientDetail.id, 'restart')}
              >
                重启
              </Button>
              <Button icon={<EditOutlined />} onClick={() => handleOpenEdit(clientDetail.id)}>
                编辑
              </Button>
            </Space>

            <Tabs
              activeKey={detailTab}
              onChange={setDetailTab}
              items={[
                {
                  key: 'overview',
                  label: '概览',
                  children: (
                    <Space direction="vertical" size="large" style={{ width: '100%' }}>
                      {statsGrid}
                      <Card size="small" title="基础配置">
                        <Descriptions column={1} size="small" bordered>
                          <Descriptions.Item label="名称">
                            {clientDetail.name}
                          </Descriptions.Item>
                          <Descriptions.Item label="描述">
                            {clientDetail.description || '-'}
                          </Descriptions.Item>
                          <Descriptions.Item label="状态">
                            {renderClientStatusTag(clientDetail.status)}
                          </Descriptions.Item>
                          <Descriptions.Item label="Smee URL">
                            <Space size={8}>
                              <a href={clientDetail.smeeUrl} target="_blank" rel="noreferrer">
                                {clientDetail.smeeUrl}
                              </a>
                              <Tooltip title="复制 Smee URL">
                                <Button
                                  size="small"
                                  type="text"
                                  icon={<CopyOutlined />}
                                  onClick={() => copyToClipboard(clientDetail.smeeUrl || '', message)}
                                />
                              </Tooltip>
                            </Space>
                          </Descriptions.Item>
                          <Descriptions.Item label="Target URL">
                            <Space size={8}>
                              <a href={clientDetail.targetUrl} target="_blank" rel="noreferrer">
                                {clientDetail.targetUrl}
                              </a>
                              <Tooltip title="复制 Target URL">
                                <Button
                                  size="small"
                                  type="text"
                                  icon={<CopyOutlined />}
                                  onClick={() => copyToClipboard(clientDetail.targetUrl || '', message)}
                                />
                              </Tooltip>
                            </Space>
                          </Descriptions.Item>
                          <Descriptions.Item label="连接超时">
                            {clientDetail.targetTimeout || 60} 秒
                          </Descriptions.Item>
                          <Descriptions.Item label="脚本格式">
                            {clientDetail.httpie ? 'HTTPie' : 'cURL'}
                          </Descriptions.Item>
                          <Descriptions.Item label="忽略事件">
                            {clientDetail.ignoreEvents?.length
                              ? clientDetail.ignoreEvents.join(', ')
                              : '无'}
                          </Descriptions.Item>
                          <Descriptions.Item label="仅保存不转发">
                            {clientDetail.noReplay ? '是' : '否'}
                          </Descriptions.Item>
                          <Descriptions.Item label="SSE 缓冲区">
                            {formatBytes(clientDetail.sseBufferSize)}
                          </Descriptions.Item>
                          <Descriptions.Item label="创建时间">
                            {formatDateTime(clientDetail.createdAt)}
                          </Descriptions.Item>
                          <Descriptions.Item label="更新时间">
                            {formatDateTime(clientDetail.updatedAt)}
                          </Descriptions.Item>
                          <Descriptions.Item label="进程信息">
                            <Space direction="vertical" size={4}>
                              <span>PID：{clientDetail.pid || '-'}</span>
                              <span>重启次数：{clientDetail.restartCount || 0}</span>
                              <span>最后错误：{clientDetail.lastError || '-'}</span>
                            </Space>
                          </Descriptions.Item>
                        </Descriptions>
                      </Card>
                    </Space>
                  ),
                },
                {
                  key: 'events',
                  label: '事件',
                  children: (
                    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                      <Space wrap>
                        <Button
                          icon={<ThunderboltOutlined />}
                          disabled={!selectedEventKeys.length}
                          onClick={() => handleReplayEvents(selectedEventKeys)}
                        >
                          批量重放
                        </Button>
                        <Button
                          onClick={() => setSelectedEventKeys([])}
                          disabled={!selectedEventKeys.length}
                        >
                          清空选择
                        </Button>
                      </Space>
                      <Table
                        rowKey="id"
                        dataSource={events}
                        columns={eventColumns}
                        loading={eventsLoading}
                        scroll={{ x: 800 }}
                        rowSelection={{
                          selectedRowKeys: selectedEventKeys,
                          onChange: setSelectedEventKeys,
                        }}
                        pagination={{
                          current: eventPagination.page,
                          pageSize: eventPagination.pageSize,
                          total: eventPagination.total,
                          showSizeChanger: true,
                          showTotal: (total) => `共 ${total} 条事件`,
                        }}
                        onChange={(pagination, filters, sorter) => {
                          setEventPagination({
                            page: pagination.current,
                            pageSize: pagination.pageSize,
                            total: eventPagination.total,
                          });
                          if (sorter && sorter.field) {
                            setEventSortState({
                              field: sorter.field,
                              order: sorter.order === 'ascend' ? 'asc' : 'desc',
                            });
                          } else {
                            setEventSortState({ field: 'timestamp', order: 'desc' });
                          }
                        }}
                      />
                    </Space>
                  ),
                },
                {
                  key: 'logs',
                  label: '日志',
                  children: (
                    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                      <Alert
                        type="info"
                        message="日志面板仅展示之后的实时日志，历史记录请查看服务端日志。"
                        showIcon
                      />
                      <Space wrap align="center">
                        <Tooltip title="启用/暂停自动滚动">
                          <Button
                            icon={<PauseCircleOutlined />}
                            type={logsAutoScroll ? 'primary' : 'default'}
                            onClick={() => setLogsAutoScroll((prev) => !prev)}
                          >
                            自动滚动
                          </Button>
                        </Tooltip>
                        <Typography.Text type="secondary">
                          已连接日志流，后续输出会实时追加。
                        </Typography.Text>
                      </Space>
                      <Card size="small">
                        <div className="logs-container" ref={logsContainerRef}>
                          {logs.length === 0 && (
                            <Typography.Text type="secondary">暂无日志输出</Typography.Text>
                          )}
                          {logs.map((line, index) => (
                            <div key={`${line}-${index}`} className="log-line">
                              {line}
                            </div>
                          ))}
                        </div>
                      </Card>
                    </Space>
                  ),
                },
              ]}
            />
          </Space>
        )}
      </Modal>

      <EventDetailModal
        open={eventDetailVisible}
        loading={eventDetailLoading}
        event={eventDetailData}
        onClose={() => {
          setEventDetailVisible(false);
          setEventDetailData(null);
        }}
        onReplay={(eventId) => handleReplayEvents([eventId])}
        onDelete={(eventId) => handleDeleteEvent(eventId)}
        messageApi={message}
      />
    </Layout>
  );
}

function App() {
  return (
    <AntApp>
      <AppContent />
    </AntApp>
  );
}

export default App;
