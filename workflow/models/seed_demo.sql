-- 重置 demo 数据：清掉 graph_id 以 g- 开头的旧记录，再灌入符合 index.html 渲染 schema 的 5 条
DELETE FROM wf_graph WHERE graph_id LIKE 'g-%';

INSERT INTO wf_graph (graph_id, name, description, graph, version, type, record_id, status, create_by, update_by) VALUES
(
  'g-order-001',
  '订单创建工作流',
  '电商下单主流程：校验用户→扣库存→生成订单',
  '{"nodes":[{"node_id":"start","name":"开始","type":"in","successors":["validate"]},{"node_id":"validate","name":"校验用户","type":"action","action_type":"validate_user","success_node_id":["stock"],"fail_node_id":["end"]},{"node_id":"stock","name":"扣减库存","type":"action","action_type":"deduct_stock","success_node_id":["create"],"fail_node_id":["end"]},{"node_id":"create","name":"生成订单","type":"action","action_type":"create_order","request_mode":"async","success_node_id":["end"]},{"node_id":"end","name":"结束","type":"out"}]}',
  1, 'svc_pipe', 1001, 2, 'alice', 'alice'
),
(
  'g-refund-002',
  '退款审批流',
  '金额>500 走经理审批；否则直接退款',
  '{"nodes":[{"node_id":"start","name":"开始","type":"in","successors":["check_amount"]},{"node_id":"check_amount","name":"金额判断","type":"condition","check_list":[{"check_param":{"rule":"amount>500"},"hit_node_id":"approve"}],"default_node_id":"refund"},{"node_id":"approve","name":"经理审批","type":"action","action_type":"manager_approve","success_node_id":["refund"],"fail_node_id":["fail_state"]},{"node_id":"refund","name":"执行退款","type":"action","action_type":"do_refund","success_node_id":["success_state"],"fail_node_id":["fail_state"]},{"node_id":"success_state","name":"退款成功","type":"state"},{"node_id":"fail_state","name":"退款拒绝","type":"state"}]}',
  2, 'flow_pipe', 1002, 2, 'bob', 'bob'
),
(
  'g-notify-003',
  '消息通知编排',
  '并行下发短信、邮件、IM',
  '{"nodes":[{"node_id":"start","name":"开始","type":"in","successors":["sms","mail","im"]},{"node_id":"sms","name":"短信","type":"action","action_type":"send_sms","request_mode":"async","success_node_id":["end"]},{"node_id":"mail","name":"邮件","type":"action","action_type":"send_mail","request_mode":"async","success_node_id":["end"]},{"node_id":"im","name":"IM","type":"action","action_type":"send_im","request_mode":"async","success_node_id":["end"]},{"node_id":"end","name":"结束","type":"out"}]}',
  1, 'svc_pipe', 1003, 1, 'carol', 'carol'
),
(
  'g-risk-004',
  '风控决策树',
  '基于评分多分支：>=80 拦截 / >=50 人工复核 / 其它通过',
  '{"nodes":[{"node_id":"start","name":"开始","type":"in","successors":["score"]},{"node_id":"score","name":"风险评分","type":"action","action_type":"calc_risk_score","success_node_id":["branch"]},{"node_id":"branch","name":"风险分级","type":"condition","check_list":[{"check_param":{"rule":"score>=80"},"hit_node_id":"blocked"},{"check_param":{"rule":"score>=50"},"hit_node_id":"review"}],"default_node_id":"approved"},{"node_id":"review","name":"人工复核","type":"action","action_type":"manual_review","success_node_id":["approved"],"fail_node_id":["blocked"]},{"node_id":"approved","name":"通过","type":"state"},{"node_id":"blocked","name":"拦截","type":"state"}]}',
  3, 'flow_pipe', 1004, 3, 'dave', 'eve'
),
(
  'g-onboard-005',
  '新员工入职流',
  '账号开通→设备发放→培训登记',
  '{"nodes":[{"node_id":"start","name":"开始","type":"in","successors":["account"]},{"node_id":"account","name":"开通账号","type":"action","action_type":"create_account","success_node_id":["device"],"fail_node_id":["end"]},{"node_id":"device","name":"发放设备","type":"action","action_type":"issue_device","success_node_id":["train"]},{"node_id":"train","name":"培训登记","type":"action","action_type":"register_training","request_mode":"async","success_node_id":["end"]},{"node_id":"end","name":"完成","type":"out"}]}',
  1, 'flow_pipe', 1005, 1, 'frank', 'frank'
);
