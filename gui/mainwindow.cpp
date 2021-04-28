#include "mainwindow.h"
#include "ui_mainwindow.h"
#include <QStandardItemModel>
#include <QTableWidget>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    QTableView *t = ui->tableView;
    QStandardItemModel *m = new QStandardItemModel;
    m->setHorizontalHeaderLabels(QStringList{"漫画标题", "章节标题", "进度"});
    t->setModel(m);
    t->setSelectionBehavior(QAbstractItemView::SelectRows);
}

MainWindow::~MainWindow()
{
    delete ui;
}
