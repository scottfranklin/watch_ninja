#include <dirent.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/inotify.h>
#include <unistd.h>

// Size of the event structure, not counting name.
#define EVENT_SIZE  (sizeof (struct inotify_event))

// Size of read buffer.
#define BUF_LEN     (128 * (EVENT_SIZE + 16))


static int is_dir(const struct dirent* entry) {
    return (entry->d_type == DT_DIR) && (entry->d_name[0] != '.');
}

static int get_dirs(char* path, char*** dirs, int* sz, int num) {
    struct dirent** entries;
    int n = scandir (path, &entries, is_dir, alphasort);

    int p_len = strlen(path);

    // Resize if necessary.
    if (num + n > *sz) {
        *sz = (num + n) * 2;
        *dirs = (char**)realloc(*dirs, sizeof(char*) * (*sz));
    }
        
    if (n >= 0) {
        for (int i = 0; i < n; i++) {
            char* d = malloc(p_len + strlen(entries[i]->d_name) + 2);
            strcpy(d, path);
            d[p_len] = '/';
            strcpy(d + p_len + 1, entries[i]->d_name);
            (*dirs)[num + i] = d;
        }
        int base = num;
        num += n;
        for (int i = 0; i < n; i++) {
            num = get_dirs((*dirs)[base + i], dirs, sz, num);
        }
        free(entries);
        return num;
    }
    free(entries);
    return 0;
}

int main(int argc, char* argv) {
    int sz = 10;
    char** dirs = malloc(sizeof(char*) * sz);
    dirs[0] = ".";
    int n = get_dirs(".", &dirs, &sz, 1);
    int fd = inotify_init();
    if (fd < 0)
        perror("inotify_init");

    char** watchers = (char**)malloc(sizeof(char*) * (n + 1));

    for (int i = 0; i < n; i++) {
        int wd = inotify_add_watch(fd, dirs[i], IN_ATTRIB);
        if (wd < 0)
            perror("inotify_add_watch");
        watchers[wd] = dirs[i];
        printf("Watching: %s/\n", dirs[i]);
    }

    char buf[BUF_LEN];
    char cmd_buf[2 * BUF_LEN];
    int len, i, j;
    fd_set rfds;

    while (1) {
        FD_ZERO(&rfds);
        FD_SET(fd, &rfds);
        FD_SET(0, &rfds);
        int retval = select(fd + 1, &rfds, NULL, NULL, NULL);
        if (retval <= 0) {
            perror("select()");
        } else {
            if (FD_ISSET(0, &rfds)) {
                // stdout.
                puts("% ninja");
                read(0, buf, BUF_LEN);
                system("ninja");
            }
            if (FD_ISSET(fd, &rfds)) {
                len = read(fd, buf, BUF_LEN);
                if (len < 0) {
                    if (errno == EINTR) {
                        printf("EINTR!\n");
                    } else {
                        perror ("read");
                    }
                    continue;
                } else if (!len) {
                    printf("len == 0!!\n");
                    continue;
                }
                strcpy(cmd_buf, "ninja ");
                i = 0; j = strlen(cmd_buf);
                int got_one = 0;
                while (i < len) {
                    struct inotify_event *event;

                    event = (struct inotify_event*)&buf[i];

                    if (event->len) {
                        char* base = watchers[event->wd];
                        int bl = strlen(base);
                        int el = strlen(event->name);

                        if (event->name[el - 1] != 'c' && (el < 4 || event->name[el - 1] != 'p' || event->name[el - 3] != 'c')) {
                            i += EVENT_SIZE + event->len;
                            continue;
                        }

                        memcpy(cmd_buf + j, base, bl);
                        j += bl;
                        cmd_buf[j + 0] = '/';
                        j++;
                        memcpy(cmd_buf + j, event->name, el);
                        j += el;
                        cmd_buf[j + 0] = '^';
                        cmd_buf[j + 1] = ' ';
                        cmd_buf[j + 2] = 0;
                        j += 2;
                        got_one = 1;
                    }
                    i += EVENT_SIZE + event->len;
                }
                if (got_one) {
                    printf("%% %s\n", cmd_buf); 
                    system(cmd_buf);
                }
            }
        }

        // Buffer more events in case lots of things are changing.
        sleep(1);
    }
}
