m4_dnl mount.m4
m4_dnl Supporting mount macros for build.mk.m4
m4_dnl By J. Stuart McMurray
m4_dnl Created 20260110
m4_dnl Last Modified 20260110
m4_dnl
m4_dnl mount mounts m4's $1 on Make's dir_$@ and writes the vnd device to
m4_dnl Make's dev_$@.  M4's $2 will be passed via mount's -o option.
m4_define(m4_mount,
	`mkdir -p $`@'_dir
	doas vnconfig ${.CURDIR}/$1 >$`@'_dev
	doas mount m4_dnl
-o `nodev,noexec,noperm,nosuid'm4_ifelse($#,1,`',`,$2') m4_dnl
/dev/$$(<$`@'_dev)a $`@'_dir 'm4_dnl
)m4_dnl
m4_dnl
m4_dnl umount umounts Make's dir_$@, releases the vnd device written to
m4_dnl Make's dev_$@, and deletes the directory and file.
m4_define(m4_umount,
	`doas umount $`@'_dir
	doas vnconfig -u $$(<$`@'_dev)
	rm $`@'_dev
	rmdir $`@'_dir'm4_dnl
)m4_dnl
m4_dnl vim: si noet
