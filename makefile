# Makefile

# path to your cluster file and TB binary:
TB_BIN     := ./data/tigerbeetle
CLUSTER_DB := ./data/0_0.tigerbeetle

.PHONY: tb-init tb-start tb-run

# 1) Remove any existing database file
tb-init:
	rm -f $(CLUSTER_DB)
	$(TB_BIN) format \
		--cluster=0 \
		--replica=0 \
		--replica-count=1 \
		--development \
		$(CLUSTER_DB)

# 2) Fire up the TB daemon (in foreground)
tb-start:
	$(TB_BIN) start \
		--addresses=0.0.0.0:3000 \
		--development \
		$(CLUSTER_DB)

# 3) Combined: init then start
tb-run: tb-init tb-start
